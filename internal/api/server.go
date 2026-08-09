// Package api serves the JSON API and the embedded frontend.
//
// The trust boundary lives here. The client posts the five picks it made
// and nothing else; the server re-derives the rolls from the season's seed,
// replays those picks through the same simulation package the browser used,
// and computes the authoritative score.
//
// What that claim is NOT: it does not make the game unsearchable. 240 legal
// drafts is trivially enumerable when the simulation runs in the browser,
// and enumerating them wins 40% of titles against 17.7% for the best
// scripted line. An earlier version of this comment said a cheater "would
// have to find decisions that genuinely produce a high score, which is just
// playing well" -- true of the continuous sliders that are long gone, and
// not true of any draft. Scores cannot be fabricated; they can be searched
// for. See docs/_README.md.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jvonmikael/lights-out/internal/scheduler"
	"github.com/jvonmikael/lights-out/internal/store"
)

// Server holds the API dependencies.
type Server struct {
	store    *store.Store
	sched    *scheduler.Scheduler
	log      *slog.Logger
	metrics  *Metrics
	registry *prometheus.Registry
	limiter  *rateLimiter
	staticFS http.Handler
	now      func() time.Time
}

// Options configure a Server.
type Options struct {
	Store     *store.Store
	Scheduler *scheduler.Scheduler
	Logger    *slog.Logger
	// Static serves the built frontend. Optional: nil disables it, which is
	// what the handler tests do.
	Static http.Handler
	// SubmitPerMinute caps submissions per client IP. Zero means 10.
	SubmitPerMinute int
}

// NewServer wires a Server.
func NewServer(o Options) *Server {
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.SubmitPerMinute == 0 {
		o.SubmitPerMinute = 10
	}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)
	if o.Scheduler != nil {
		o.Scheduler.OnPublish = metrics.SeasonsPublished.Inc
	}
	return &Server{
		store:    o.Store,
		sched:    o.Scheduler,
		log:      o.Logger,
		metrics:  metrics,
		registry: reg,
		limiter:  newRateLimiter(o.SubmitPerMinute, time.Minute),
		staticFS: o.Static,
		now:      time.Now,
	}
}

// Handler builds the router. Go 1.22's ServeMux does method and path
// matching, so a router dependency is no longer justified at this scale.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/seasons/today", s.observe("/api/seasons/today", s.handleTodaySeason))
	mux.HandleFunc("GET /api/seasons/{id}/leaderboard", s.observe("/api/seasons/{id}/leaderboard", s.handleLeaderboard))
	mux.HandleFunc("POST /api/runs", s.observe("/api/runs", s.handleSubmitRun))
	mux.HandleFunc("GET /healthz", s.observe("/healthz", s.handleHealthz))
	mux.Handle("GET /metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))

	if s.staticFS != nil {
		mux.Handle("GET /", s.staticFS)
	}
	return mux
}
