-- Visibility enum for quilts:
--   - 'private' means only the owning user can see it.
--   - 'public' means its allowed to be visible/linked to from the app.
--   - 'protected' means anyone who knows the URL can see it, but it's not
--     advertised.
CREATE TYPE visibility AS ENUM ('public', 'private', 'protected');

-- Main table for quilts.
CREATE TABLE quilts (
    quilt_id    SERIAL      PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(user_id),
    name        VARCHAR     NOT NULL,
    visibility  visibility  NOT NULL DEFAULT 'private',
    width       INTEGER     NOT NULL CHECK (width > 0),
    height      INTEGER     NOT NULL CHECK (width > 0),
    created     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    modified    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

-- Table for quilt polygons.
CREATE TABLE quilt_polys (
    quilt_poly_id SERIAL  PRIMARY KEY,
    quilt_id      INTEGER NOT NULL REFERENCES quilts(quilt_id),
    fabric_id     INTEGER NOT NULL REFERENCES fabrics DEFAULT fabric_color('ffffff'),
    poly          geometry(POLYGON) NOT NULL
);

-- Function that evaluates the quilt_polys table for a particular quilt
-- to make sure there are no overlapping polygons in that quilt.
CREATE OR REPLACE FUNCTION check_no_overlapping_polys(
    _quilt_poly_id INTEGER, _quilt_id INTEGER, _poly geometry(POLYGON))
RETURNS BOOLEAN AS $$
DECLARE
    result BOOLEAN;
BEGIN
    -- Detecting overlap:
    --   ST_Disjoint(p1,p2) returns true if p1 and p2 are completely separate
    --   ST_Touches(p1,p2) returns true if p1 and p2 only share edges (this is
    --     find for our quilt polys)
    -- The following query computes the boolean and of whether all polys in
    -- _quilt_id (other than _quilt_poly_id) are either disjoint from or only
    -- touch _poly.
    SELECT bool_and(ST_Disjoint(_poly, poly) OR ST_Touches(_poly, poly))
        INTO result
        FROM quilt_polys
        WHERE quilt_id = _quilt_id AND quilt_poly_id != _quilt_poly_id;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Add the above function as a check constraint on the quilt_polys table.
ALTER TABLE quilt_polys ADD CONSTRAINT no_overlapping_polys CHECK
    (check_no_overlapping_polys(quilt_poly_id, quilt_id, poly));
