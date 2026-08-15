package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jvonmikael/lights-out/internal/sim"
)

// newTestStore brings up a real Postgres in a container, runs the real
// migrations against it, and returns a connected Store. No mocks, no
// sqlmock, no manual setup -- docs/Tech Stack.md is explicit that the
// database under test should be a real one.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	// Wait for the readiness log line TWICE. The postgres image starts a
	// temporary server to run initdb and then restarts, so port 5432 is
	// listening about a second in, long before the database will complete a
	// handshake. Waiting on the port alone connects mid-init and the client
	// hangs rather than failing.
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("lightsout"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("starting postgres: %v", err)
	}
	t.Cleanup(func() {
		if err := pg.Terminate(context.Background()); err != nil {
			t.Logf("terminating container: %v", err)
		}
	})

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	s, err := New(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(s.Close)

	if err := s.Migrate(); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}
	return s
}

func testSeason(t *testing.T, s *Store, seed int64) int64 {
	t.Helper()
	season := sim.GenerateSeason(seed)
	cal, _ := json.Marshal(season.Calendar)
	field, _ := json.Marshal(season.Rivals)
	row, err := s.CreateSeason(context.Background(), CreateSeasonParams{
		Seed:       seed,
		SimVersion: sim.Version,
		Calendar:   cal,
		Field:      field,
	})
	if err != nil {
		t.Fatalf("creating season: %v", err)
	}
	return row.ID
}

func TestPingAndMigrate(t *testing.T) {
	s := newTestStore(t)
	if err := s.Ping(context.Background()); err != nil {
		t.Errorf("ping after migrate: %v", err)
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	s := newTestStore(t)
	// Running migrations again on an already-migrated database must be a
	// no-op, because the server calls Migrate on every startup.
	if err := s.Migrate(); err != nil {
		t.Errorf("second Migrate: %v", err)
	}
}

func TestSeasonRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	id := testSeason(t, s, 4242)

	got, err := s.SeasonByID(ctx, id)
	if err != nil {
		t.Fatalf("SeasonByID: %v", err)
	}
	if got.ID != id {
		t.Errorf("got season %d, want %d", got.ID, id)
	}
	if got.Seed != 4242 {
		t.Errorf("seed = %d, want 4242", got.Seed)
	}
	if got.SimVersion != sim.Version {
		t.Errorf("sim version = %q, want %q", got.SimVersion, sim.Version)
	}
	if got.CreatedAt.IsZero() {
		t.Error("created_at was not populated")
	}
}

// Under unlimited play, issuing a season must ALWAYS mint a new row. The
// old schema had UNIQUE (published_at) to make a daily scheduler
// idempotent, which is precisely the constraint that would now stop a
// player having a second go.
func TestEveryIssuedSeasonIsANewRow(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ids := map[int64]bool{}
	for i := 0; i < 5; i++ {
		id := testSeason(t, s, int64(i))
		if ids[id] {
			t.Fatalf("season id %d issued twice", id)
		}
		ids[id] = true
	}

	// Even the SAME seed twice: the server mints seeds, so a collision is
	// possible and must not be an error.
	a, b := testSeason(t, s, 4242), testSeason(t, s, 4242)
	if a == b {
		t.Error("the same seed twice produced one row; a second run would be blocked")
	}

	n, err := s.CountSeasons(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Errorf("%d seasons exist, want 7", n)
	}
}

func TestOneRunPerPlayerPerSeason(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seasonID := testSeason(t, s, 7)

	playerID := uuid.New()
	if _, err := s.UpsertPlayer(ctx, playerID, "alice"); err != nil {
		t.Fatal(err)
	}

	picks := sim.Strategy("bestavailable", sim.GenerateSeason(7))
	if picks == nil {
		t.Fatal("unknown strategy name -- the strategy list changed")
	}
	res, err := sim.RunSeason(7, picks)
	if err != nil {
		t.Fatal(err)
	}
	decisions, _ := json.Marshal(picks)
	full, _ := json.Marshal(res)
	params := SaveRunParams{
		SeasonID: seasonID, PlayerID: playerID, Decisions: decisions,
		Points: res.Player.Points, Wins: res.Player.Wins,
		Podiums: res.Player.Podiums, DNFs: res.Player.DNFs, Result: full,
	}

	if _, err := s.SaveRun(ctx, params); err != nil {
		t.Fatalf("first submission: %v", err)
	}

	// The second must be rejected by the DATABASE, not by application logic
	// that can race between the check and the insert.
	_, err = s.SaveRun(ctx, params)
	if !errors.Is(err, ErrAlreadySubmitted) {
		t.Fatalf("second submission returned %v, want ErrAlreadySubmitted", err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code != "23505" {
		t.Errorf("underlying error is %s, want unique violation 23505", pgErr.Code)
	}
}

func TestLeaderboardOrdering(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seasonID := testSeason(t, s, 11)

	// Points descending, then wins, then podiums, then earliest submission.
	entries := []struct {
		name                        string
		points, wins, podiums, dnfs int
	}{
		{"carol", 200, 5, 8, 0},
		{"alice", 250, 3, 7, 1},
		{"dave", 200, 5, 6, 2},
		{"bob", 200, 7, 9, 0},
	}
	for _, e := range entries {
		id := uuid.New()
		if _, err := s.UpsertPlayer(ctx, id, e.name); err != nil {
			t.Fatal(err)
		}
		if _, err := s.SaveRun(ctx, SaveRunParams{
			SeasonID: seasonID, PlayerID: id, Decisions: []byte(`[]`),
			Points: e.points, Wins: e.wins, Podiums: e.podiums, DNFs: e.dnfs,
			Result: []byte(`{}`),
		}); err != nil {
			t.Fatal(err)
		}
	}

	board, err := s.Leaderboard(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alice", "bob", "carol", "dave"}
	if len(board) != len(want) {
		t.Fatalf("leaderboard has %d rows, want %d", len(board), len(want))
	}
	for i, name := range want {
		if board[i].DisplayName != name {
			t.Errorf("rank %d is %q, want %q", i+1, board[i].DisplayName, name)
		}
		if board[i].Rank != i+1 {
			t.Errorf("row %d has rank %d", i, board[i].Rank)
		}
	}
}

func TestLeaderboardPaginates(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seasonID := testSeason(t, s, 13)

	for i := 0; i < 5; i++ {
		id := uuid.New()
		s.UpsertPlayer(ctx, id, "p")
		if _, err := s.SaveRun(ctx, SaveRunParams{
			SeasonID: seasonID, PlayerID: id, Decisions: []byte(`[]`),
			Points: 100 - i, Result: []byte(`{}`),
		}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.Leaderboard(ctx, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 2 {
		t.Fatalf("page has %d rows, want 2", len(page))
	}
	if page[0].Rank != 3 || page[1].Rank != 4 {
		t.Errorf("ranks are %d,%d -- want 3,4 (offset must carry into rank)", page[0].Rank, page[1].Rank)
	}
}

func TestUpsertPlayerUpdatesName(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	id := uuid.New()
	if _, err := s.UpsertPlayer(ctx, id, "first"); err != nil {
		t.Fatal(err)
	}
	p, err := s.UpsertPlayer(ctx, id, "second")
	if err != nil {
		t.Fatal(err)
	}
	if p.DisplayName != "second" {
		t.Errorf("display name = %q, want %q", p.DisplayName, "second")
	}
}

func TestRunForPlayerReportsMissing(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seasonID := testSeason(t, s, 17)

	_, err := s.RunForPlayer(ctx, seasonID, uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("missing run returned %v, want ErrNotFound", err)
	}
}

func TestSeasonByIDReportsMissing(t *testing.T) {
	s := newTestStore(t)
	_, err := s.SeasonByID(context.Background(), 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("missing season returned %v, want ErrNotFound", err)
	}
}

// One row per player, their best run, plus how many they played. Under
// unlimited play a best-of-N is partly a measure of N, so the board reports
// the N rather than hiding it.
func TestLeaderboardShowsEachPlayersBestAndTheirRunCount(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id := uuid.New()
	if _, err := s.UpsertPlayer(ctx, id, "grinder"); err != nil {
		t.Fatal(err)
	}
	for i, points := range []int{40, 260, 120} {
		if _, err := s.SaveRun(ctx, SaveRunParams{
			SeasonID: testSeason(t, s, int64(500+i)), PlayerID: id,
			Decisions: []byte(`[]`), Points: points, Result: []byte(`{}`),
		}); err != nil {
			t.Fatal(err)
		}
	}

	board, err := s.Leaderboard(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(board) != 1 {
		t.Fatalf("board has %d rows for one player, want 1", len(board))
	}
	if board[0].Points != 260 {
		t.Errorf("board shows %d points, want the best of 40/260/120", board[0].Points)
	}
	if board[0].Runs != 3 {
		t.Errorf("board shows %d runs, want 3", board[0].Runs)
	}
}
