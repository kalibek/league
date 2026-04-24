CREATE TYPE event_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE league_events (
    event_id     BIGSERIAL PRIMARY KEY,
    league_id    BIGINT NOT NULL REFERENCES leagues(league_id),
    status       event_status NOT NULL DEFAULT 'DRAFT',
    title        VARCHAR(255) NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);
