-- Unlimited play replaces the daily seed.
--
-- A season is no longer a day. It is one issued run: the server mints a
-- seed on request, records it, and verifies the picks that come back
-- against it. Nothing about that is time-bounded, so published_at and
-- closes_at have no meaning left -- and UNIQUE (published_at) actively
-- prevented a second season existing on the same day, which is exactly
-- what unlimited play needs.
ALTER TABLE seasons DROP CONSTRAINT seasons_published_at_key;
ALTER TABLE seasons DROP COLUMN published_at;
ALTER TABLE seasons DROP COLUMN closes_at;
ALTER TABLE seasons ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();

-- The leaderboard is all-time now rather than per-season, so it is no
-- longer prefixed by season_id and needs its own index.
--
-- UNIQUE (season_id, player_id) on runs is UNCHANGED and still does its
-- job: one submission per issued season. Play again and you are issued a
-- new one, which is the whole point.
DROP INDEX runs_leaderboard;
CREATE INDEX runs_alltime ON runs (points DESC, wins DESC, podiums DESC, created_at ASC);
CREATE INDEX runs_by_player ON runs (player_id);
