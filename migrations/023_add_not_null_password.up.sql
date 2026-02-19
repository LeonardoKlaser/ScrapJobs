-- Set empty passwords to a placeholder hash before adding NOT NULL
UPDATE users SET user_password = '$2a$10$placeholder' WHERE user_password IS NULL;
ALTER TABLE users ALTER COLUMN user_password SET NOT NULL;
