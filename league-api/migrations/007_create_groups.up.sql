CREATE TYPE group_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE groups (
    group_id     BIGSERIAL PRIMARY KEY,
    event_id     BIGINT NOT NULL REFERENCES league_events(event_id),
    status       group_status NOT NULL DEFAULT 'DRAFT',
    division     VARCHAR(10) NOT NULL,
    group_no     INT NOT NULL,
    scheduled    TIMESTAMP NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);
