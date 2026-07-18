-- Schema as specified in docs/Architecture.md, "Data model".

CREATE TABLE seasons (
  id           bigserial PRIMARY KEY,
  seed         bigint      NOT NULL,
  sim_version  text        NOT NULL,
  calendar     jsonb       NOT NULL,   -- circuits, profiles, budgets
  field        jsonb       NOT NULL,   -- rival teams, starting ratings, AI archetypes
  published_at date        NOT NULL,
  closes_at    timestamptz NOT NULL,
  UNIQUE (published_at)
);

CREATE TABLE players (
  id           uuid PRIMARY KEY,       -- generated client-side, stored in localStorage
  display_name text        NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE runs (
  id          bigserial PRIMARY KEY,
  season_id   bigint      NOT NULL REFERENCES seasons(id),
  player_id   uuid        NOT NULL REFERENCES players(id),
  decisions   jsonb       NOT NULL,    -- the only thing the client supplies
  points      int         NOT NULL,    -- computed server-side
  wins        int         NOT NULL,
  podiums     int         NOT NULL,
  dnfs        int         NOT NULL,
  result      jsonb       NOT NULL,    -- full race-by-race, for the share card
  created_at  timestamptz NOT NULL DEFAULT now(),
  -- The entire anti-resubmission mechanism: one run per player per season,
  -- enforced by the database rather than by application logic that can race.
  UNIQUE (season_id, player_id)
);

CREATE INDEX runs_leaderboard ON runs (season_id, points DESC, wins DESC, podiums DESC);
