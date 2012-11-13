-- Enum for the two subclasses of fabrics.
CREATE TYPE fabric_type AS ENUM ('image', 'color');

-- Table for fabrics supertype.
-- If fabric_type is image, FIXME.
-- If fabric_type is color, an entry must exist for fabric_id in fabric_colors.
CREATE TABLE fabrics (
    fabric_id   SERIAL      PRIMARY KEY,
    fabric_type fabric_type NOT NULL
);

-- Solid color fabrics. The color attribute must be a 6-character hex color
-- code as would be suitable for HTML.
CREATE TABLE fabric_colors (
    fabric_id   INTEGER PRIMARY KEY REFERENCES fabrics(fabric_id),
    color       CHAR(6) CONSTRAINT check_color
        CHECK (color IS NOT NULL AND color ~ '^[a-fA-F0-9]{6}$'),
    UNIQUE (color)
);

-- Helper function to create a colored fabric and its associated entry in
-- the fabrics table.
CREATE OR REPLACE FUNCTION fabric_color(_color CHAR(6))
RETURNS INTEGER AS $$
DECLARE
    id INTEGER;
BEGIN
    LOOP
        SELECT fabric_id INTO id FROM fabric_colors WHERE color=_color;
        IF FOUND THEN
            RETURN id;
        END IF;

        BEGIN
            INSERT INTO fabrics(fabric_type) VALUES('color')
                RETURNING fabric_id INTO id;
            INSERT INTO fabric_colors(fabric_id, color) VALUES(id, _color);
        EXCEPTION WHEN unique_violation THEN
            -- ignore and go back into loop to select the id
        END;
        RETURN id;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
