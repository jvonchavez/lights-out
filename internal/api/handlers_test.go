package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jvonmikael/lights-out/internal/scheduler"
	"github.com/jvonmikael/lights-out/internal/sim"
	"github.com/jvonmikael/lights-out/internal/store"
)

type testEnv struct {
	t        *testing.T
	srv      *Server
	handler  http.Handler
	store    *store.Store
	seasonID int64
	seed     int64
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("lightsout"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("starting postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	if err := st.Migrate(); err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	sched := scheduler.New(st, log)
	srv := NewServer(Options{Store: st, Scheduler: sched, Logger: log, SubmitPerMinute: 1000})

	season, err := sched.PublishToday(ctx)
	if err != nil {
		t.Fatalf("publishing season: %v", err)
	}

	return &testEnv{
		t: t, srv: srv, handler: srv.Handler(), store: st,
		seasonID: season.ID, seed: season.Seed,
	}
}

func (e *testEnv) do(method, path, body string) *httptest.ResponseRecorder {
	e.t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, r)
	return rec
}

// middlePicksJSON is a legal pick for every window.
func middlePicksJSON() string {
	parts := make([]string, sim.WindowCount)
	for i := range parts {
		parts[i] = "1"
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func middlePicks() []int {
	p := make([]int, sim.WindowCount)
	for i := range p {
		p[i] = 1
	}
	return p
}

// TestForgedScoreIsIgnored is the M3 definition of done.
//
// The submission carries points, score, wins and a whole fabricated result
// object claiming a perfect season. The server must ignore every one of
// them and score the decisions instead.
func TestForgedScoreIsIgnored(t *testing.T) {
	env := newTestEnv(t)

	body := fmt.Sprintf(`{
		"season_id": %d,
		"player_id": %q,
		"display_name": "cheater",
		"points": 999999,
		"score": 999999,
		"wins": 10,
		"podiums": 10,
		"dnfs": 0,
		"position": 1,
		"rank": 1,
		"result": {"player": {"points": 999999, "wins": 10}},
		"build": [{"name": "Fabricated Part"}],
		"picks": %s
	}`, env.seasonID, uuid.NewString(), middlePicksJSON())

	rec := env.do("POST", "/api/runs", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var got submitResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	want, err := sim.RunSeason(env.seed, middlePicks())
	if err != nil {
		t.Fatal(err)
	}

	if got.Points == 999999 {
		t.Fatal("the forged score was accepted")
	}
	if got.Points != want.Player.Points {
		t.Errorf("scored %d points, want %d -- the forged value leaked through",
			got.Points, want.Player.Points)
	}
	if got.Wins != want.Player.Wins {
		t.Errorf("wins = %d, want %d", got.Wins, want.Player.Wins)
	}
	if got.DNFs != want.Player.DNFs {
		t.Errorf("dnfs = %d, want %d", got.DNFs, want.Player.DNFs)
	}
	if got.Share != want.Share {
		t.Errorf("share = %q, want %q", got.Share, want.Share)
	}

	// And what was PERSISTED must be the computed score too, not just what
	// was echoed back in the response.
	board, err := env.store.Leaderboard(context.Background(), env.seasonID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(board) != 1 {
		t.Fatalf("leaderboard has %d rows, want 1", len(board))
	}
	if board[0].Points != want.Player.Points {
		t.Errorf("persisted %d points, want %d", board[0].Points, want.Player.Points)
	}
}

func TestSubmissionValidation(t *testing.T) {
	env := newTestEnv(t)

	shortPicks := make([]string, sim.WindowCount-1)
	for i := range shortPicks {
		shortPicks[i] = "1"
	}

	tests := []struct {
		name string
		body string
		want int
	}{
		{"malformed json", `{`, http.StatusBadRequest},
		{"bad player id", fmt.Sprintf(`{"season_id":%d,"player_id":"not-a-uuid","picks":%s}`,
			env.seasonID, middlePicksJSON()), http.StatusBadRequest},
		{"too few picks", fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":[%s]}`,
			env.seasonID, uuid.NewString(), strings.Join(shortPicks, ",")), http.StatusBadRequest},
		{"negative card index", fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":[-1,1,1,1,1]}`,
			env.seasonID, uuid.NewString()), http.StatusBadRequest},
		{"card index past the deal", fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":[0,0,9,0,0]}`,
			env.seasonID, uuid.NewString()), http.StatusBadRequest},
		{"unknown season", fmt.Sprintf(`{"season_id":999999,"player_id":%q,"picks":%s}`,
			uuid.NewString(), middlePicksJSON()), http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := env.do("POST", "/api/runs", tt.body)
			if rec.Code != tt.want {
				t.Errorf("status %d, want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestOneSubmissionPerPlayerPerSeason(t *testing.T) {
	env := newTestEnv(t)
	player := uuid.NewString()
	body := fmt.Sprintf(`{"season_id":%d,"player_id":%q,"display_name":"alice","picks":%s}`,
		env.seasonID, player, middlePicksJSON())

	if rec := env.do("POST", "/api/runs", body); rec.Code != http.StatusCreated {
		t.Fatalf("first submission: %d %s", rec.Code, rec.Body.String())
	}
	rec := env.do("POST", "/api/runs", body)
	if rec.Code != http.StatusConflict {
		t.Errorf("second submission: status %d, want 409: %s", rec.Code, rec.Body.String())
	}
}

func TestClosedSeasonRejectsSubmission(t *testing.T) {
	env := newTestEnv(t)
	// Jump the server's clock past the season's close.
	env.srv.now = func() time.Time { return time.Now().UTC().Add(48 * time.Hour) }

	body := fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":%s}`,
		env.seasonID, uuid.NewString(), middlePicksJSON())
	rec := env.do("POST", "/api/runs", body)
	if rec.Code != http.StatusConflict {
		t.Errorf("status %d, want 409 for a closed season: %s", rec.Code, rec.Body.String())
	}
}

func TestRateLimitOnSubmission(t *testing.T) {
	env := newTestEnv(t)
	env.srv.limiter = newRateLimiter(2, time.Minute)

	codes := make([]int, 0, 4)
	for i := 0; i < 4; i++ {
		body := fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":%s}`,
			env.seasonID, uuid.NewString(), middlePicksJSON())
		codes = append(codes, env.do("POST", "/api/runs", body).Code)
	}
	if codes[0] != http.StatusCreated || codes[1] != http.StatusCreated {
		t.Errorf("first two submissions were %v, want 201s", codes[:2])
	}
	if codes[2] != http.StatusTooManyRequests || codes[3] != http.StatusTooManyRequests {
		t.Errorf("submissions 3 and 4 were %v, want 429s", codes[2:])
	}
}

func TestTodaySeasonEndpoint(t *testing.T) {
	env := newTestEnv(t)
	rec := env.do("GET", "/api/seasons/today", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var got seasonResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != env.seasonID {
		t.Errorf("season id = %d, want %d", got.ID, env.seasonID)
	}
	if got.Seed != fmt.Sprint(env.seed) {
		t.Errorf("seed = %q, want %q", got.Seed, fmt.Sprint(env.seed))
	}
	// The seed must be a JSON string, not a number, or JS loses precision.
	var raw map[string]json.RawMessage
	json.Unmarshal(rec.Body.Bytes(), &raw)
	if !strings.HasPrefix(string(raw["seed"]), `"`) {
		t.Errorf("seed serialised as %s, want a JSON string", raw["seed"])
	}
	if got.SimVersion != sim.Version {
		t.Errorf("sim version = %q, want %q", got.SimVersion, sim.Version)
	}
	for w, deal := range got.Deals {
		for i, c := range deal {
			if c.ID == "" || c.Name == "" {
				t.Errorf("window %d card %d came back empty", w, i)
			}
		}
	}
	if got.WindowRounds != sim.WindowRounds {
		t.Errorf("window rounds = %v, want %v", got.WindowRounds, sim.WindowRounds)
	}
}

func TestLeaderboardEndpoint(t *testing.T) {
	env := newTestEnv(t)
	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"season_id":%d,"player_id":%q,"display_name":"p%d","picks":%s}`,
			env.seasonID, uuid.NewString(), i, middlePicksJSON())
		if rec := env.do("POST", "/api/runs", body); rec.Code != http.StatusCreated {
			t.Fatalf("submission %d: %d", i, rec.Code)
		}
	}
	rec := env.do("GET", fmt.Sprintf("/api/seasons/%d/leaderboard", env.seasonID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Entries []store.LeaderboardEntry `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 3 {
		t.Fatalf("%d entries, want 3", len(got.Entries))
	}
	for i, e := range got.Entries {
		if e.Rank != i+1 {
			t.Errorf("entry %d has rank %d", i, e.Rank)
		}
	}
}

func TestLeaderboardForUnknownSeasonIsEmpty(t *testing.T) {
	env := newTestEnv(t)
	rec := env.do("GET", "/api/seasons/424242/leaderboard", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"entries":[]`) {
		t.Errorf("body = %s, want an empty entries array", rec.Body.String())
	}
}

func TestHealthzAndMetrics(t *testing.T) {
	env := newTestEnv(t)

	if rec := env.do("GET", "/healthz", ""); rec.Code != http.StatusOK {
		t.Errorf("healthz: %d %s", rec.Code, rec.Body.String())
	}

	body := fmt.Sprintf(`{"season_id":%d,"player_id":%q,"picks":%s}`,
		env.seasonID, uuid.NewString(), middlePicksJSON())
	env.do("POST", "/api/runs", body)

	rec := env.do("GET", "/metrics", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: %d", rec.Code)
	}
	for _, want := range []string{
		"http_request_duration_seconds",
		"sim_verification_duration_seconds",
		`runs_submitted_total{outcome="accepted"}`,
	} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("/metrics is missing %q", want)
		}
	}
}

func TestRequestIDHeaderIsSet(t *testing.T) {
	env := newTestEnv(t)
	rec := env.do("GET", "/api/seasons/today", "")
	if id := rec.Header().Get("X-Request-Id"); id == "" {
		t.Error("no X-Request-Id header")
	} else if _, err := uuid.Parse(id); err != nil {
		t.Errorf("X-Request-Id %q is not a UUID", id)
	}
}

func TestSeasonsPublishedCounterIncrements(t *testing.T) {
	env := newTestEnv(t)

	// Measure the delta rather than an absolute value: newTestEnv already
	// publishes one season, and asserting a fixed count couples this test
	// to that setup detail.
	before := publishedCount(t, env)
	env.srv.sched.OnPublish()
	after := publishedCount(t, env)

	if after != before+1 {
		t.Errorf("seasons_published_total went %v -> %v, want +1", before, after)
	}
	if before < 1 {
		t.Errorf("counter was %v before the test; publishing a season must increment it", before)
	}
}

// publishedCount scrapes seasons_published_total from /metrics.
func publishedCount(t *testing.T, env *testEnv) float64 {
	t.Helper()
	body := env.do("GET", "/metrics", "").Body.String()
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "seasons_published_total ") {
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "seasons_published_total ")), 64)
			if err != nil {
				t.Fatalf("parsing %q: %v", line, err)
			}
			return v
		}
	}
	t.Fatalf("/metrics has no seasons_published_total:\n%s", grepLines(body, "seasons"))
	return 0
}

func grepLines(body, needle string) string {
	var out []string
	for _, l := range strings.Split(body, "\n") {
		if strings.Contains(l, needle) {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
