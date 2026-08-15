// Package store owns everything that touches Postgres: the connection pool,
// the embedded migrations, and a small hand-written surface over the
// sqlc-generated queries.
//
// The wrapper exists to translate database-specific failures into domain
// errors the API layer can map to status codes without importing pgx. A
// unique violation on (season_id, player_id) is not a 500; it means the
// player already submitted today.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver, for migrations only

	"github.com/jvonmikael/lights-out/internal/store/db"
)

// Migrations are embedded so the binary carries its own schema and can
// migrate on startup with no files on disk. embed cannot reach outside its
// own package directory, which is why they live here rather than in a
// top-level db/ folder.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

var (
	// ErrNotFound means the row does not exist.
	ErrNotFound = errors.New("store: not found")
	// ErrAlreadySubmitted means this player already has a run for this
	// season. The API maps it to 409.
	ErrAlreadySubmitted = errors.New("store: player already submitted for this season")
)

// Store is the database handle.
type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
	dsn  string
}

// New opens a connection pool and verifies it with a ping.
func New(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("store: parsing dsn: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	cfg.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("store: connecting: %w", err)
	}
	// Bound the first ping. Without a deadline a half-open connection --
	// a Postgres still running initdb, say -- blocks forever instead of
	// failing, which turns a clear error into a mystery hang.
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool, q: db.New(pool), dsn: dsn}, nil
}

// Close releases the pool.
func (s *Store) Close() { s.pool.Close() }

// Ping reports whether Postgres is reachable. /healthz is gated on it.
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// Migrate applies any outstanding migrations. It is safe to call on every
// startup: golang-migrate is a no-op when the schema is already current.
func (s *Store) Migrate() error {
	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("store: reading migrations: %w", err)
	}

	// Migrations run on their OWN connection, not the application pool.
	// golang-migrate holds a connection for its advisory lock for the whole
	// migration; borrowing that from the pool leaves it checked out, and
	// pool.Close() then blocks forever waiting for a connection that is
	// never returned.
	sqlDB, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("store: opening migration connection: %w", err)
	}
	defer sqlDB.Close()

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("store: migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("store: migrator: %w", err)
	}
	// Releases the advisory lock and the driver's connection.
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("store: applying migrations: %w", err)
	}
	return nil
}

// Season is one issued run: a seed the server minted, with the calendar and
// field it produces. It is no longer a day -- see migration 000002.
type Season struct {
	ID         int64
	Seed       int64
	SimVersion string
	Calendar   []byte
	Field      []byte
	CreatedAt  time.Time
}

// CreateSeasonParams are the fields needed to issue a season.
type CreateSeasonParams struct {
	Seed       int64
	SimVersion string
	Calendar   []byte
	Field      []byte
}

func seasonFromRow(r db.Season) Season {
	return Season{
		ID: r.ID, Seed: r.Seed, SimVersion: r.SimVersion,
		Calendar: r.Calendar, Field: r.Field,
		CreatedAt: r.CreatedAt.Time,
	}
}

// CreateSeason issues a season. Every call mints a new one -- there is no
// conflict to handle, because unlimited play means a player can be issued
// as many as they ask for.
//
// This is the ONLY place a seed enters the system, which is what keeps the
// server authoritative: a client cannot nominate a seed it has already
// solved offline.
func (s *Store) CreateSeason(ctx context.Context, p CreateSeasonParams) (Season, error) {
	row, err := s.q.CreateSeason(ctx, db.CreateSeasonParams{
		Seed:       p.Seed,
		SimVersion: p.SimVersion,
		Calendar:   p.Calendar,
		Field:      p.Field,
	})
	if err != nil {
		return Season{}, fmt.Errorf("store: creating season: %w", err)
	}
	return seasonFromRow(row), nil
}

// SeasonByID returns one season.
func (s *Store) SeasonByID(ctx context.Context, id int64) (Season, error) {
	row, err := s.q.GetSeasonByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Season{}, ErrNotFound
	}
	if err != nil {
		return Season{}, fmt.Errorf("store: reading season: %w", err)
	}
	return seasonFromRow(row), nil
}

// CountSeasons is used by tests.
func (s *Store) CountSeasons(ctx context.Context) (int64, error) {
	var n int64
	err := s.pool.QueryRow(ctx, "SELECT count(*) FROM seasons").Scan(&n)
	return n, err
}

// Player is a registered player. Identity is a client-generated UUID; there
// are no passwords, sessions, or PII anywhere in this table.
type Player struct {
	ID          uuid.UUID
	DisplayName string
}

// UpsertPlayer registers a player or updates their display name.
func (s *Store) UpsertPlayer(ctx context.Context, id uuid.UUID, name string) (Player, error) {
	row, err := s.q.UpsertPlayer(ctx, db.UpsertPlayerParams{ID: id, DisplayName: name})
	if err != nil {
		return Player{}, fmt.Errorf("store: upserting player: %w", err)
	}
	return Player{ID: row.ID, DisplayName: row.DisplayName}, nil
}

// SaveRunParams is a verified run ready to persist. Every scoring field is
// computed server-side; the client supplies only Decisions.
type SaveRunParams struct {
	SeasonID  int64
	PlayerID  uuid.UUID
	Decisions []byte
	Points    int
	Wins      int
	Podiums   int
	DNFs      int
	Result    []byte
}

// Run is a persisted run.
type Run struct {
	ID       int64
	SeasonID int64
	PlayerID uuid.UUID
	Points   int
	Wins     int
	Podiums  int
	DNFs     int
	Result   []byte
}

func runFromRow(r db.Run) Run {
	return Run{
		ID: r.ID, SeasonID: r.SeasonID, PlayerID: r.PlayerID,
		Points: int(r.Points), Wins: int(r.Wins),
		Podiums: int(r.Podiums), DNFs: int(r.Dnfs), Result: r.Result,
	}
}

// SaveRun persists a verified run, returning ErrAlreadySubmitted when this
// player already has one for this season. The uniqueness is enforced by the
// database rather than a read-then-write in application code, which would
// race between two concurrent submissions.
func (s *Store) SaveRun(ctx context.Context, p SaveRunParams) (Run, error) {
	row, err := s.q.CreateRun(ctx, db.CreateRunParams{
		SeasonID:  p.SeasonID,
		PlayerID:  p.PlayerID,
		Decisions: p.Decisions,
		Points:    int32(p.Points),
		Wins:      int32(p.Wins),
		Podiums:   int32(p.Podiums),
		Dnfs:      int32(p.DNFs),
		Result:    p.Result,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Wrapped, so callers can still reach the PgError with
			// errors.As while matching on the domain error.
			return Run{}, fmt.Errorf("%w: %w", ErrAlreadySubmitted, err)
		}
		return Run{}, fmt.Errorf("store: saving run: %w", err)
	}
	return runFromRow(row), nil
}

// RunForPlayer returns a player's run for a season.
func (s *Store) RunForPlayer(ctx context.Context, seasonID int64, playerID uuid.UUID) (Run, error) {
	row, err := s.q.GetRunForPlayer(ctx, db.GetRunForPlayerParams{SeasonID: seasonID, PlayerID: playerID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ErrNotFound
	}
	if err != nil {
		return Run{}, fmt.Errorf("store: reading run: %w", err)
	}
	return runFromRow(row), nil
}

// LeaderboardEntry is one ranked row: a player's best season, and how many
// they have played.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	PlayerID    uuid.UUID `json:"player_id"`
	DisplayName string    `json:"display_name"`
	Points      int       `json:"points"`
	Wins        int       `json:"wins"`
	Podiums     int       `json:"podiums"`
	DNFs        int       `json:"dnfs"`
	Runs        int       `json:"runs"`
}

// Leaderboard returns a page of the all-time standings -- one row per
// player, their best season -- ranked by points, then wins, then podiums,
// then earliest submission. Rank accounts for the offset so a second page
// continues the numbering.
//
// Runs comes back with each row on purpose. When play is unlimited a
// best-of-N is partly a measure of N, and showing the N is the honest way
// to present that.
func (s *Store) Leaderboard(ctx context.Context, limit, offset int32) ([]LeaderboardEntry, error) {
	rows, err := s.q.GetLeaderboard(ctx, db.GetLeaderboardParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("store: reading leaderboard: %w", err)
	}
	out := make([]LeaderboardEntry, 0, len(rows))
	for i, r := range rows {
		out = append(out, LeaderboardEntry{
			Rank:        int(offset) + i + 1,
			PlayerID:    r.PlayerID,
			DisplayName: r.DisplayName,
			Points:      int(r.Points),
			Wins:        int(r.Wins),
			Podiums:     int(r.Podiums),
			DNFs:        int(r.Dnfs),
			Runs:        int(r.Runs),
		})
	}
	return out, nil
}
