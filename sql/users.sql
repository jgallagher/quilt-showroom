-- Main table for users. Rules represented here:
--   - Displayed names and emails must be globally unique.
--   - Displayed names must contain only letters, numbers, _, and -.
CREATE TABLE users (
    user_id     SERIAL  PRIMARY KEY,
    name        VARCHAR UNIQUE  CONSTRAINT check_name
        CHECK (name IS NOT NULL AND name ~ '^[-a-zA-Z0-9_]+$'),
    email       VARCHAR UNIQUE  NOT NULL,
    password    VARCHAR         NOT NULL
);

-- Possible outcomes of the users_create function:
--  - success
--  - dup_name - the supplied name already exists
--  - dup_email - the supplied email aready exists
--  - bad_name - the name fails our restrictions on displayed names
--  - bad_pass - the password fails our requirements
-- Note that there is no 'bad_email' code, as verifying email addresses is
-- difficult to the point of uselessness.
CREATE TYPE users_create_code AS ENUM (
    'success', 'dup_name', 'dup_email', 'bad_name', 'bad_pass');
CREATE TYPE users_create_return AS (
    code users_create_code,
    user_id INT
);

-- Stored procedure to handle creating a new user. If successful, returns
-- (success,new_user_id). If unsuccessful, returns (CODE, NULL) where CODE
-- is one of the values above.
CREATE OR REPLACE FUNCTION users_create(
    name VARCHAR, email VARCHAR, password VARCHAR)
RETURNS users_create_return AS $$
DECLARE
    user_id users.user_id%TYPE;
    ret users_create_return;
BEGIN
    -- See if the requested username already exists.
    IF EXISTS (SELECT 1 FROM users WHERE users.name = users_create.name) THEN
        ret.code = 'dup_name';
        RETURN ret;
    END IF;

    -- See if the requested email already exists.
    IF EXISTS (SELECT 1 FROM users WHERE users.email = users_create.email) THEN
        ret.code = 'dup_email';
        RETURN ret;
    END IF;

    -- Everything looks good - insert and return the new user_id.
    -- Note that we use a BEGIN ... WHEN to catch two exceptions:
    --  - raise_exception may be thrown by our qlt_users_password_trigger
    --  - check_violation may be thrown by the check constraint on the name
    --    column
    BEGIN
        INSERT INTO users(name, email, password) VALUES(name, email, password)
            RETURNING users.user_id INTO ret.user_id;
    EXCEPTION
        WHEN raise_exception THEN
            ret.code = 'bad_pass';
            RETURN ret;
        WHEN check_violation THEN
            ret.code = 'bad_name';
            RETURN ret;
    END;

    -- No exception on insert - success!
    ret.code = 'success';
    RETURN ret;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to handle user passwords:
--  - Meet our website password requirements (at least 8 characters long).
--  - Perform password hashing.
CREATE OR REPLACE FUNCTION trigger_users_password()
RETURNS trigger AS $$
BEGIN
    -- Check password length.
    IF LENGTH(NEW.password) < 8 THEN
        RAISE EXCEPTION 'Password must be at least 8 characters long.';
    END IF;

    -- Hash passwords.
    NEW.password = crypt(NEW.password, gen_salt('bf'));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to the table.
CREATE TRIGGER trigger_users_password BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE trigger_users_password();
