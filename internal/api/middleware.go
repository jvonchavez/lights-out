package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const requestIDKey ctxKey = iota

// statusRecorder captures the status code for logging and metrics.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// RequestIDFrom returns the request ID carried on a context, if any.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// observe assigns a request ID, logs the request, records its duration, and
// turns a panic into a 500 rather than a dropped connection.
func (s *Server) observe(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := uuid.NewString()
		w.Header().Set("X-Request-Id", id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey, id))

		rec := &statusRecorder{ResponseWriter: w}

		defer func() {
			if v := recover(); v != nil {
				s.log.Error("panic serving request",
					"request_id", id, "route", route, "panic", v)
				if rec.status == 0 {
					rec.status = http.StatusInternalServerError
					w.WriteHeader(rec.status)
				}
			}
			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			dur := time.Since(start)
			s.metrics.RequestDuration.
				WithLabelValues(route, r.Method, strconv.Itoa(rec.status)).
				Observe(dur.Seconds())
			s.log.Info("request",
				"request_id", id, "method", r.Method, "route", route,
				"status", rec.status, "duration_ms", dur.Milliseconds())
		}()

		next(rec, r)
	}
}

// rateLimiter is a per-IP token bucket. It guards submission only: reads
// are cheap, but a submission costs a full native season simulation.
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity int
	window   time.Duration
	now      func() time.Time
}

type bucket struct {
	tokens int
	resets time.Time
}

func newRateLimiter(capacity int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		window:   window,
		now:      time.Now,
	}
}

// allow reports whether this key may proceed, consuming a token if so.
func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()

	b, ok := l.buckets[key]
	if !ok || now.After(b.resets) {
		l.buckets[key] = &bucket{tokens: l.capacity - 1, resets: now.Add(l.window)}
		// Opportunistically drop expired buckets so the map cannot grow
		// without bound across a long uptime.
		if len(l.buckets) > 10000 {
			for k, v := range l.buckets {
				if now.After(v.resets) {
					delete(l.buckets, k)
				}
			}
		}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// clientIP prefers the left-most X-Forwarded-For entry, since App Runner
// terminates TLS and proxies.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return trimSpace(xff[:i])
			}
		}
		return trimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

var _ = slog.Default
