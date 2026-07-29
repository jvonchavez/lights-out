package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/jvonmikael/lights-out/internal/scheduler"
	"github.com/jvonmikael/lights-out/internal/sim"
	"github.com/jvonmikael/lights-out/internal/store"
)

// writeJSON emits a JSON body with a status code.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.log.Error("writing response", "error", err)
	}
}

type errorBody struct {
	Error string `json:"error"`
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, errorBody{Error: msg})
}

// seasonResponse is the descriptor the client needs to render the game.
//
// Seed is a STRING. JavaScript numbers are float64, and although seeds are
// masked below 2^53 the string keeps the contract explicit at the boundary
// rather than relying on that invariant holding forever.
type seasonResponse struct {
	ID         int64           `json:"id"`
	Seed       string          `json:"seed"`
	SimVersion string          `json:"sim_version"`
	Calendar   json.RawMessage `json:"calendar"`
	Field      json.RawMessage `json:"field"`
	// Deals are re-derived from the seed rather than stored, so the client
	// can render cards before the WASM module finishes loading and the
	// server is always the authority on what was offered.
	Deals        [sim.WindowCount][sim.DealSize]sim.Card `json:"deals"`
	WindowRounds [sim.WindowCount]int                    `json:"window_rounds"`
	ClosesAt     time.Time                               `json:"closes_at"`
}

func seasonToResponse(row store.Season) seasonResponse {
	return seasonResponse{
		ID:           row.ID,
		Seed:         strconv.FormatInt(row.Seed, 10),
		SimVersion:   row.SimVersion,
		Calendar:     row.Calendar,
		Field:        row.Field,
		Deals:        sim.DealsFor(row.Seed),
		WindowRounds: sim.WindowRounds,
		ClosesAt:     row.ClosesAt,
	}
}

// handleTodaySeason returns today's season, publishing it if the scheduler
// has not yet done so. Publishing here as well as on the ticker means a
// cold start serves a season on the first request rather than a 404.
func (s *Server) handleTodaySeason(w http.ResponseWriter, r *http.Request) {
	day := scheduler.Day(s.now())
	row, err := s.store.SeasonByDate(r.Context(), day)
	if errors.Is(err, store.ErrNotFound) {
		row, err = s.sched.PublishToday(r.Context())
	}
	if err != nil {
		s.log.Error("loading today's season", "request_id", RequestIDFrom(r.Context()), "error", err)
		s.writeError(w, http.StatusInternalServerError, "could not load today's season")
		return
	}
	s.writeJSON(w, http.StatusOK, seasonToResponse(row))
}

// submitRequest is everything the client may supply.
//
// There is no score field, no points field, and no result field. A client
// that sends one is not rejected -- the value simply has nowhere to be
// decoded to, and is discarded by encoding/json. That is the cleanest
// possible implementation of the trust boundary: the forged number cannot
// be read, so it cannot be believed.
//
// Picks are card indices, so the client cannot even describe a development
// that was not offered. The deal is re-derived from the season's seed.
type submitRequest struct {
	SeasonID    int64  `json:"season_id"`
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	Picks       []int  `json:"picks"`
}

type submitResponse struct {
	Rank     int              `json:"rank"`
	Points   int              `json:"points"`
	Wins     int              `json:"wins"`
	Podiums  int              `json:"podiums"`
	DNFs     int              `json:"dnfs"`
	Position int              `json:"position"`
	Share    string           `json:"share"`
	Result   sim.SeasonResult `json:"result"`
}

func (s *Server) handleSubmitRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := RequestIDFrom(ctx)

	if !s.limiter.allow(clientIP(r)) {
		s.metrics.RunsSubmitted.WithLabelValues("rate_limited").Inc()
		s.writeError(w, http.StatusTooManyRequests, "too many submissions, try again shortly")
		return
	}

	var req submitRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	if err := dec.Decode(&req); err != nil {
		s.metrics.RunsSubmitted.WithLabelValues("bad_request").Inc()
		s.writeError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		s.metrics.RunsSubmitted.WithLabelValues("bad_request").Inc()
		s.writeError(w, http.StatusBadRequest, "player_id must be a UUID")
		return
	}
	name := trimSpace(req.DisplayName)
	if name == "" {
		name = "Anonymous"
	}
	if len(name) > 32 {
		name = name[:32]
	}

	season, err := s.store.SeasonByID(ctx, req.SeasonID)
	if errors.Is(err, store.ErrNotFound) {
		s.metrics.RunsSubmitted.WithLabelValues("no_season").Inc()
		s.writeError(w, http.StatusNotFound, "no such season")
		return
	}
	if err != nil {
		s.log.Error("loading season", "request_id", reqID, "error", err)
		s.writeError(w, http.StatusInternalServerError, "could not load season")
		return
	}

	if !s.now().Before(season.ClosesAt) {
		s.metrics.RunsSubmitted.WithLabelValues("closed").Inc()
		s.writeError(w, http.StatusConflict, "this season is closed")
		return
	}

	// THE VERIFICATION. The season's seed comes from the database, never
	// from the request, and the score is computed here from the submitted
	// decisions. Anything the client could lie about is recomputed.
	start := time.Now()
	result, err := sim.RunSeason(season.Seed, req.Picks)
	verifyDur := time.Since(start)
	s.metrics.VerifyDuration.Observe(verifyDur.Seconds())
	if err != nil {
		s.metrics.RunsSubmitted.WithLabelValues("invalid_decisions").Inc()
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := s.store.UpsertPlayer(ctx, playerID, name); err != nil {
		s.log.Error("upserting player", "request_id", reqID, "error", err)
		s.writeError(w, http.StatusInternalServerError, "could not record player")
		return
	}

	decisions, err := json.Marshal(req.Picks)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "could not encode picks")
		return
	}
	full, err := json.Marshal(result)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "could not encode result")
		return
	}

	if _, err := s.store.SaveRun(ctx, store.SaveRunParams{
		SeasonID:  season.ID,
		PlayerID:  playerID,
		Decisions: decisions,
		Points:    result.Player.Points,
		Wins:      result.Player.Wins,
		Podiums:   result.Player.Podiums,
		DNFs:      result.Player.DNFs,
		Result:    full,
	}); err != nil {
		if errors.Is(err, store.ErrAlreadySubmitted) {
			s.metrics.RunsSubmitted.WithLabelValues("duplicate").Inc()
			s.writeError(w, http.StatusConflict, "you have already submitted for this season")
			return
		}
		s.log.Error("saving run", "request_id", reqID, "error", err)
		s.writeError(w, http.StatusInternalServerError, "could not save run")
		return
	}

	s.metrics.RunsSubmitted.WithLabelValues("accepted").Inc()
	s.log.Info("run verified",
		"request_id", reqID, "season_id", season.ID, "player_id", playerID,
		"points", result.Player.Points, "verify_ms", verifyDur.Milliseconds(),
		"outcome", "accepted")

	rank := 0
	if board, err := s.store.Leaderboard(ctx, season.ID, 1000, 0); err == nil {
		for _, e := range board {
			if e.PlayerID == playerID {
				rank = e.Rank
				break
			}
		}
	}

	s.writeJSON(w, http.StatusCreated, submitResponse{
		Rank:     rank,
		Points:   result.Player.Points,
		Wins:     result.Player.Wins,
		Podiums:  result.Player.Podiums,
		DNFs:     result.Player.DNFs,
		Position: result.PlayerPos,
		Share:    result.Share,
		Result:   result,
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "season id must be an integer")
		return
	}
	limit := clampQueryInt(r, "limit", 50, 1, 200)
	offset := clampQueryInt(r, "offset", 0, 0, 100000)

	board, err := s.store.Leaderboard(r.Context(), id, int32(limit), int32(offset))
	if err != nil {
		s.log.Error("reading leaderboard", "request_id", RequestIDFrom(r.Context()), "error", err)
		s.writeError(w, http.StatusInternalServerError, "could not load leaderboard")
		return
	}
	if board == nil {
		board = []store.LeaderboardEntry{}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"entries": board})
}

func clampQueryInt(r *http.Request, key string, def, lo, hi int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// handleHealthz returns 200 only after a successful Postgres ping.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		s.writeError(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok", "sim_version": sim.Version,
	})
}
