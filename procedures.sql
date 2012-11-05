DROP TYPE IF EXISTS qlt_create_user_return CASCADE;
DROP TYPE IF EXISTS qlt_create_user_code CASCADE;

-- Possible outcomes of the qlt_create_user function:
--  - success
--  - dup_name - the supplied name already exists
--  - dup_email - the supplied email aready exists
--  - bad_name - the name fails our restrictions on displayed names
--  - bad_pass - the password fails our requirements
-- Note that there is no 'bad_email' code, as verifying email addresses is
-- difficult to the point of uselessness.
CREATE TYPE qlt_create_user_code AS ENUM (
    'success', 'dup_name', 'dup_email', 'bad_name', 'bad_pass');
CREATE TYPE qlt_create_user_return AS (
    code qlt_create_user_code,
    user_id INT
);

CREATE OR REPLACE FUNCTION qlt_create_user(name VARCHAR,
    email VARCHAR, password VARCHAR)
RETURNS qlt_create_user_return AS $$
DECLARE
    user_id users.user_id%TYPE;
    ret qlt_create_user_return;
BEGIN
    -- See if the requested username already exists.
    IF EXISTS (SELECT 1 FROM users WHERE users.name = qlt_create_user.name) THEN
        ret.code = 'dup_name';
        RETURN ret;
    END IF;

    -- See if the requested email already exists.
    IF EXISTS (SELECT 1 FROM users WHERE users.email = qlt_create_user.email) THEN
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
CREATE OR REPLACE FUNCTION qlt_users_password_trigger()
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
CREATE TRIGGER qlt_users_password_trigger BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE qlt_users_password_trigger();
