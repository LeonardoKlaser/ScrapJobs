CREATE TABLE requested_sites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    url VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);