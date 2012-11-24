-- Main table for blocks.
CREATE TABLE blocks (
    block_id    SERIAL  PRIMARY KEY,
    user_id     VARCHAR NOT NULL REFERENCES users(user_id),
    name        VARCHAR NOT NULL,
    width       INTEGER NOT NULL CHECK (width > 0),
    height      INTEGER NOT NULL CHECK (height > 0)
);

-- Table for block polygons.
CREATE TABLE block_polys (
    block_poly_id   SERIAL  PRIMARY KEY,
    block_id        INTEGER NOT NULL REFERENCES blocks(block_id),
    fabric_id       INTEGER NOT NULL REFERENCES fabrics(fabric_id)
                        DEFAULT fabric_color('ffffff'),
    poly            geometry(POLYGON) NOT NULL
);

-- Function that evaluates the block_polys table for a particular block to
-- make sure there are no overlapping polygons in that block.
-- This is functionally equivalent to quilt_no_overlapping_polys in quilts.sql.
CREATE OR REPLACE FUNCTION block_no_overlapping_polys(
    _block_poly_id INTEGER, _block_id INTEGER, _poly geometry(POLYGON))
RETURNS BOOLEAN AS $$
DECLARE
    result BOOLEAN;
BEGIN
    SELECT bool_and(ST_Disjoint(_poly, poly) OR ST_Touches(_poly, poly))
        INTO result
        FROM block_polys
        WHERE block_id = _block_id AND block_poly_id != _block_poly_id;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Function that evaluates a polygon in the block_polys to make sure it fits
-- inside its block boundaries. All blocks start at (0,0) and extend to
-- (width, height) from their row in the blocks table.
-- This is functionally equivalent to polygon_inside_quilt in quilts.sql.
CREATE OR REPLACE FUNCTION polygon_inside_block
    (_block_id INTEGER, _poly geometry(POLYGON))
RETURNS BOOLEAN AS $$
DECLARE
    result BOOLEAN;
BEGIN
    SELECT ST_Contains(
        ST_MakeBox2D(ST_Point(0,0), ST_Point(width, height)), _poly)
    INTO result
    FROM blocks WHERE block_id = _block_id;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Add the above functions as check constraints on the block_polys table.
ALTER TABLE block_polys ADD CONSTRAINT no_overlapping_polys CHECK
    (block_no_overlapping_polys(block_poly_id, block_id, poly));
ALTER TABLE block_polys ADD CONSTRAINT poly_inside_block CHECK
    (polygon_inside_block(block_id, poly));
