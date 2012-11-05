DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name TEXT UNIQUE CONSTRAINT check_name
        CHECK (name IS NOT NULL AND name ~ '^[a-zA-Z0-9_]+$'),
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);
