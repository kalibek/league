CREATE TABLE leagues (
    league_id     BIGSERIAL PRIMARY KEY,
    title         VARCHAR(255) NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    configuration JSONB NOT NULL DEFAULT '{}',
    created       TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated  TIMESTAMP NOT NULL DEFAULT NOW()
);
