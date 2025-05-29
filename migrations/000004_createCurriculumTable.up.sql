CREATE TABLE IF NOT EXISTS curriculum(
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experience JSONB NOT NULL,
    education JSONB NOT NULL,
    skills VARCHAR NOT NULL,
    languages VARCHAR NOT NULL,
    summary TEXT NOT NULL,
)

CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR NOT NULL UNIQUE,
    password VARCHAR,
    curriculum_id INTEGER REFERENCES curriculum(id) ON DELETE CASCADE,
)
