CREATE TABLE email_provider_config (
    id SERIAL PRIMARY KEY,
    provider_name VARCHAR(20) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by INTEGER REFERENCES users(id)
);

INSERT INTO email_provider_config (provider_name, is_active, priority)
VALUES ('resend', true, 1), ('ses', true, 2);
