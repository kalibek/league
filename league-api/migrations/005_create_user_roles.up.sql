CREATE TABLE user_roles (
    user_id      BIGINT NOT NULL REFERENCES users(user_id),
    role_id      INT NOT NULL REFERENCES roles(role_id),
    league_id    BIGINT NOT NULL REFERENCES leagues(league_id),
    created      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id, league_id)
);
