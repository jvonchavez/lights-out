package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics are the four series docs/Architecture.md names. Verification
// duration is the one worth watching: it is CPU-bound work sitting on the
// request path, and if it creeps toward hundreds of milliseconds the answer
// is to move verification onto a worker queue. Instrumenting it from the
// first commit is what makes that a measurement rather than a guess.
type Metrics struct {
	RequestDuration  *prometheus.HistogramVec
	VerifyDuration   prometheus.Histogram
	RunsSubmitted    *prometheus.CounterVec
	SeasonsPublished prometheus.Counter
}

// NewMetrics registers the collectors on a registry.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	f := promauto.With(reg)
	return &Metrics{
		RequestDuration: f.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration by route and status.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method", "status"}),
		VerifyDuration: f.NewHistogram(prometheus.HistogramOpts{
			Name:    "sim_verification_duration_seconds",
			Help:    "Time spent re-running a submitted season natively.",
			Buckets: []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
		}),
		RunsSubmitted: f.NewCounterVec(prometheus.CounterOpts{
			Name: "runs_submitted_total",
			Help: "Run submissions by outcome.",
		}, []string{"outcome"}),
		SeasonsPublished: f.NewCounter(prometheus.CounterOpts{
			Name: "seasons_published_total",
			Help: "Seasons issued to players.",
		}),
	}
}
