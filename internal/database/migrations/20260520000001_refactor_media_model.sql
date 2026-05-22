-- migrate:up

-- 1. Extend storage.media with canonical key + metadata columns
ALTER TABLE storage.media
  ADD COLUMN IF NOT EXISTS storage_key text,
  ADD COLUMN IF NOT EXISTS public_url  text,
  ADD COLUMN IF NOT EXISTS alt_text    text,
  ADD COLUMN IF NOT EXISTS width       int,
  ADD COLUMN IF NOT EXISTS height      int;

-- 2. Backfill storage_key from legacy directory + filename columns
UPDATE storage.media
  SET storage_key = directory || '/' || filename
  WHERE storage_key IS NULL;

-- 3. Enforce storage_key as the single immutable bucket-path primitive
ALTER TABLE storage.media
  ALTER COLUMN storage_key SET NOT NULL;

ALTER TABLE storage.media
  DROP CONSTRAINT IF EXISTS media_disk_directory_filename_extension;

ALTER TABLE storage.media
  ADD CONSTRAINT media_storage_key_unique UNIQUE (storage_key);

-- 4. Polymorphic attachment table — links any media record to any entity+slot
CREATE TABLE IF NOT EXISTS storage.media_attachments (
  id          uuid        PRIMARY KEY DEFAULT uuidv7(),
  media_id    uuid        NOT NULL REFERENCES storage.media(id) ON DELETE CASCADE,
  entity_type text        NOT NULL,
  entity_id   uuid        NOT NULL,
  slot        text        NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (entity_type, entity_id, slot)
);

-- 5. Replace featured_image_key (text) with a proper FK on blog.posts
ALTER TABLE blog.posts
  ADD COLUMN IF NOT EXISTS featured_image_id uuid REFERENCES storage.media(id) ON DELETE SET NULL;

UPDATE blog.posts p
  SET featured_image_id = m.id
  FROM storage.media m
  WHERE m.storage_key = p.featured_image_key
    AND p.featured_image_key IS NOT NULL;

ALTER TABLE blog.posts
  DROP COLUMN IF EXISTS featured_image_key;

-- migrate:down

ALTER TABLE blog.posts
  ADD COLUMN IF NOT EXISTS featured_image_key text;

UPDATE blog.posts p
  SET featured_image_key = m.storage_key
  FROM storage.media m
  WHERE m.id = p.featured_image_id;

ALTER TABLE blog.posts
  DROP COLUMN IF EXISTS featured_image_id;

DROP TABLE IF EXISTS storage.media_attachments;

ALTER TABLE storage.media
  DROP CONSTRAINT IF EXISTS media_storage_key_unique;

ALTER TABLE storage.media
  ADD CONSTRAINT media_disk_directory_filename_extension
    UNIQUE (disk, directory, filename, extension);

ALTER TABLE storage.media DROP COLUMN IF EXISTS height;
ALTER TABLE storage.media DROP COLUMN IF EXISTS width;
ALTER TABLE storage.media DROP COLUMN IF EXISTS alt_text;
ALTER TABLE storage.media DROP COLUMN IF EXISTS public_url;
ALTER TABLE storage.media DROP COLUMN IF EXISTS storage_key;
