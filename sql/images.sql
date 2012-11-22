-- Main table for images uploaded by users.
CREATE TABLE images (
    image_id    SERIAL  PRIMARY KEY,
    user_id     VARCHAR NOT NULL REFERENCES users(user_id),
    url         VARCHAR NOT NULL
);

-- Table linking images to quilts (e.g., an uploaded picture of a completed
-- quilt).
CREATE TABLE quilt_images (
    quilt_id    INTEGER REFERENCES quilts(quilt_id),
    image_id    INTEGER REFERENCES images(image_id),
    comment     TEXT DEFAULT NULL,
    PRIMARY KEY (quilt_id, image_id)
);

-- Fabrics made out of images.
CREATE TABLE fabric_images (
    fabric_id   INTEGER PRIMARY KEY REFERENCES fabrics(fabric_id),
    image_id    INTEGER NOT NULL    REFERENCES images(image_id),
    name        VARCHAR DEFAULT NULL,
    UNIQUE (image_id)
);

-- Helper function to create an image fabric and its associated entry in
-- the fabrics table.
CREATE OR REPLACE FUNCTION fabric_image(_image_id INTEGER, _name VARCHAR)
RETURNS INTEGER AS $$
DECLARE
    id INTEGER;
BEGIN
    INSERT INTO fabrics(fabric_type) VALUES ('image')
        RETURNING fabric_id INTO id;
    INSERT INTO fabric_images(fabric_id, image_id, name)
        VALUES (id, _image_id, _name);
    INSERT INTO user_fabrics(user_id, fabric_id)
        VALUES((SELECT user_id FROM images WHERE image_id = _image_id), id);
    RETURN id;
END;
$$ LANGUAGE plpgsql;
