// Package scheduler publishes one season per UTC day.
//
// Deliberately an in-process goroutine rather than a separate service, a
// Lambda, or a cron container: seasons are a deterministic function of the
// date, so publishing is idempotent and two instances racing produce the
// same row. The UNIQUE (published_at) constraint settles the tie
// harmlessly, which is what makes this safe to run on several instances
// with no leader election.
package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"log/slog"
	"time"

	"github.com/jvonmikael/lights-out/internal/sim"
	"github.com/jvonmikael/lights-out/internal/store"
)

// maxSafeInt is 2^53 - 1. Seeds are masked into this range because they
// cross into JavaScript, where numbers are float64 and anything larger
// silently loses precision -- the browser would then play a different
// season from the one the server verifies.
const maxSafeInt = int64(1)<<53 - 1

// SeedForDate derives a season seed from a UTC date. It is a pure function
// of the date, so any instance computes the same seed for the same day.
func SeedForDate(t time.Time) int64 {
	h := fnv.New64a()
	// FNV-1a over the RFC 3339 date. Formatting explicitly rather than
	// hashing a time value keeps the seed independent of clock precision
	// and location.
	h.Write([]byte(t.UTC().Format("2006-01-02")))
	return int64(h.Sum64() & uint64(maxSafeInt)) //nolint:gosec // masked into range
}

// Day returns the UTC midnight of t.
func Day(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// Scheduler publishes daily seasons.
type Scheduler struct {
	store *store.Store
	log   *slog.Logger
	now   func() time.Time // injectable so tests need no wall clock

	// OnPublish is called when a season is newly published, and not when
	// one already existed. A callback rather than a metrics dependency so
	// this package stays free of Prometheus.
	OnPublish func()
}

// New returns a Scheduler.
func New(s *store.Store, log *slog.Logger) *Scheduler {
	return &Scheduler{store: s, log: log, now: time.Now}
}

// PublishToday ensures a season exists for the current UTC day and returns
// it. It is safe to call concurrently and repeatedly.
func (s *Scheduler) PublishToday(ctx context.Context) (store.Season, error) {
	day := Day(s.now())
	seed := SeedForDate(day)
	season := sim.GenerateSeason(seed)

	calendar, err := json.Marshal(season.Calendar)
	if err != nil {
		return store.Season{}, err
	}
	field, err := json.Marshal(season.Teams)
	if err != nil {
		return store.Season{}, err
	}

	row, err := s.store.CreateSeason(ctx, store.CreateSeasonParams{
		Seed:       seed,
		SimVersion: sim.Version,
		Calendar:   calendar,
		Field:      field,
		Day:        day,
		ClosesAt:   day.Add(24 * time.Hour),
	})
	switch {
	case err == nil:
		s.log.Info("season published",
			"season_id", row.ID, "seed", seed, "day", day.Format("2006-01-02"),
			"sim_version", sim.Version)
		if s.OnPublish != nil {
			s.OnPublish()
		}
		return row, nil
	case errors.Is(err, store.ErrAlreadyPublished):
		// Another instance won the race, or this one already ran. Both are
		// success: read back the season that exists.
		return s.store.SeasonByDate(ctx, day)
	default:
		return store.Season{}, err
	}
}

// Run publishes today's season immediately, then once an hour until ctx is
// cancelled. Hourly rather than at midnight so a restart at any time of day
// still converges quickly.
func (s *Scheduler) Run(ctx context.Context) {
	if _, err := s.PublishToday(ctx); err != nil {
		s.log.Error("publishing season at startup", "error", err)
	}
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if _, err := s.PublishToday(ctx); err != nil {
				s.log.Error("publishing season", "error", err)
			}
		}
	}
}
