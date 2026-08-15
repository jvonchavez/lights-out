DROP INDEX runs_by_player;
DROP INDEX runs_alltime;
CREATE INDEX runs_leaderboard ON runs (season_id, points DESC, wins DESC, podiums DESC);

ALTER TABLE seasons DROP COLUMN created_at;
-- The down migration cannot invent the day a season belonged to, so it
-- backfills distinct dates from the id to satisfy the unique constraint.
ALTER TABLE seasons ADD COLUMN published_at date;
ALTER TABLE seasons ADD COLUMN closes_at timestamptz;
UPDATE seasons SET published_at = DATE '2000-01-01' + id, closes_at = now();
ALTER TABLE seasons ALTER COLUMN published_at SET NOT NULL;
ALTER TABLE seasons ALTER COLUMN closes_at SET NOT NULL;
ALTER TABLE seasons ADD CONSTRAINT seasons_published_at_key UNIQUE (published_at);
