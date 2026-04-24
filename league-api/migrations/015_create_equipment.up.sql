CREATE TABLE blades (
    blade_id SERIAL PRIMARY KEY,
    name     TEXT NOT NULL UNIQUE
);

CREATE TABLE rubbers (
    rubber_id SERIAL PRIMARY KEY,
    name      TEXT NOT NULL UNIQUE
);
