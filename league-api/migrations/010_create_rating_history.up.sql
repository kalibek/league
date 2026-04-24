CREATE TABLE rating_history (
    history_id  BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(user_id),
    match_id    BIGINT NOT NULL REFERENCES matches(match_id),
    delta       DOUBLE PRECISION NOT NULL,
    rating      DOUBLE PRECISION NOT NULL,
    deviation   DOUBLE PRECISION NOT NULL,
    volatility  DOUBLE PRECISION NOT NULL
);
