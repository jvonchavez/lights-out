-- name: GetSeasonByDate :one
SELECT * FROM seasons WHERE published_at = $1;

-- name: GetSeasonByID :one
SELECT * FROM seasons WHERE id = $1;

-- name: CreateSeason :one
-- ON CONFLICT DO NOTHING is what makes the daily scheduler idempotent and
-- safe to run on several instances without leader election. It returns NO
-- ROW on conflict, so the caller must fall back to re-reading.
INSERT INTO seasons (seed, sim_version, calendar, field, published_at, closes_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (published_at) DO NOTHING
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
SELECT r.player_id, r.points, r.wins, r.podiums, r.dnfs, r.created_at, p.display_name
FROM runs r
JOIN players p ON p.id = r.player_id
WHERE r.season_id = $1
ORDER BY r.points DESC, r.wins DESC, r.podiums DESC, r.created_at ASC
LIMIT $2 OFFSET $3;

-- name: CountRuns :one
SELECT count(*) FROM runs WHERE season_id = $1;
