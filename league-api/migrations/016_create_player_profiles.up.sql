CREATE TABLE player_profiles (
    profile_id   BIGSERIAL PRIMARY KEY,
    user_id      BIGINT  NOT NULL UNIQUE REFERENCES users(user_id),
    country_id   INTEGER REFERENCES countries(country_id),
    city_id      INTEGER REFERENCES cities(city_id),
    birthdate    DATE,
    grip         TEXT CHECK (grip IN ('penhold', 'shakehand')),
    gender       TEXT CHECK (gender IN ('male', 'female', 'other')),
    blade_id     INTEGER REFERENCES blades(blade_id),
    fh_rubber_id INTEGER REFERENCES rubbers(rubber_id),
    bh_rubber_id INTEGER REFERENCES rubbers(rubber_id),
    created      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
