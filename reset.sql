-- Tables from blocks.sql.
DROP TABLE IF EXISTS block_polys CASCADE;
DROP TABLE IF EXISTS blocks CASCADE;

-- Tables from images.sql.
DROP TABLE IF EXISTS fabric_images CASCADE;
DROP TABLE IF EXISTS quilt_images CASCADE;
DROP TABLE IF EXISTS images CASCADE;

-- Tables and types from quilts.sql.
DROP TABLE IF EXISTS quilt_polys CASCADE;
DROP TABLE IF EXISTS quilt_comments CASCADE;
DROP TABLE IF EXISTS quilts CASCADE;
DROP TYPE IF EXISTS visibility CASCADE;

-- Tables and types from fabrics.sql.
DROP TABLE IF EXISTS user_fabrics CASCADE;
DROP TABLE IF EXISTS fabric_colors CASCADE;
DROP TABLE IF EXISTS fabrics CASCADE;
DROP TYPE IF EXISTS fabric_type CASCADE;

-- Tables and types from users.sql.
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS users_create_return CASCADE;
DROP TYPE IF EXISTS users_create_code CASCADE;
