CREATE TABLE group_players (
    group_player_id   BIGSERIAL PRIMARY KEY,
    group_id          BIGINT NOT NULL REFERENCES groups(group_id),
    user_id           BIGINT NOT NULL REFERENCES users(user_id),
    seed              SMALLINT NOT NULL DEFAULT 0,
    place             SMALLINT NOT NULL DEFAULT 0,
    points            SMALLINT NOT NULL DEFAULT 0,
    tiebreak_points   SMALLINT NOT NULL DEFAULT 0,
    advances          BOOLEAN NOT NULL DEFAULT FALSE,
    recedes           BOOLEAN NOT NULL DEFAULT FALSE,
    is_non_calculated BOOLEAN NOT NULL DEFAULT FALSE,
    created           TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated      TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, user_id)
);
