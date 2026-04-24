CREATE TABLE users (
    user_id        BIGSERIAL PRIMARY KEY,
    first_name     VARCHAR(100) NOT NULL,
    last_name      VARCHAR(100) NOT NULL,
    email          VARCHAR(255) NOT NULL UNIQUE,
    current_rating DOUBLE PRECISION NOT NULL DEFAULT 1500,
    deviation      DOUBLE PRECISION NOT NULL DEFAULT 350,
    volatility     DOUBLE PRECISION NOT NULL DEFAULT 0.06,
    created        TIMESTAMP NOT NULL DEFAULT NOW(),
    last_updated   TIMESTAMP NOT NULL DEFAULT NOW()
);
