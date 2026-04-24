CREATE TYPE match_status AS ENUM ('DRAFT', 'IN_PROGRESS', 'DONE');
CREATE TABLE matches (
    match_id          BIGSERIAL PRIMARY KEY,
    group_id          BIGINT NOT NULL REFERENCES groups(group_id),
    group_player1_id  BIGINT REFERENCES group_players(group_player_id),
    group_player2_id  BIGINT REFERENCES group_players(group_player_id),
    score1            SMALLINT,
    score2            SMALLINT,
    withdraw1         BOOLEAN NOT NULL DEFAULT FALSE,
    withdraw2         BOOLEAN NOT NULL DEFAULT FALSE,
    status            match_status NOT NULL DEFAULT 'DRAFT',
    created           TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated      TIMESTAMP NOT NULL DEFAULT NOW()
);
