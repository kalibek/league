CREATE TABLE cities (
    city_id    SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    country_id INTEGER NOT NULL REFERENCES countries(country_id),
    UNIQUE (name, country_id)
);
