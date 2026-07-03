DROP INDEX IF EXISTS idx_shorten_url_user_id;
ALTER TABLE shorten_url DROP COLUMN IF EXISTS user_id;