-- name: GetSeasonByID :one
SELECT * FROM seasons WHERE id = $1;

-- name: CreateSeason :one
-- One row per issued run. There is no conflict to handle any more: every
-- call mints a new season, which is what unlimited play means.
INSERT INTO seasons (seed, sim_version, calendar, field)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpsertPlayer :one
INSERT INTO players (id, display_name) VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET display_name = EXCLUDED.display_name
RETURNING *;

-- name: CreateRun :one
INSERT INTO runs (season_id, player_id, decisions, points, wins, podiums, dnfs, result)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetRunForPlayer :one
SELECT * FROM runs WHERE season_id = $1 AND player_id = $2;

-- name: GetLeaderboard :many
-- All-time, one row per player: their best season, plus how many they have
-- submitted. The run count is not decoration -- when play is unlimited a
-- best-of-N is partly a measure of N, and showing it is the honest way to
-- present that rather than pretending otherwise.
--
-- DISTINCT ON needs its own ordering to pick the best row per player, so
-- the outer query re-sorts into leaderboard order.
SELECT b.player_id, b.points, b.wins, b.podiums, b.dnfs, b.created_at, b.runs, p.display_name
FROM (
  SELECT DISTINCT ON (player_id)
         player_id, points, wins, podiums, dnfs, created_at,
         count(*) OVER (PARTITION BY player_id) AS runs
  FROM runs
  ORDER BY player_id, points DESC, wins DESC, podiums DESC, created_at ASC
) b
JOIN players p ON p.id = b.player_id
ORDER BY b.points DESC, b.wins DESC, b.podiums DESC, b.created_at ASC
LIMIT $1 OFFSET $2;

-- name: CountRuns :one
SELECT count(*) FROM runs;

-- name: CountRunsForPlayer :one
SELECT count(*) FROM runs WHERE player_id = $1;
