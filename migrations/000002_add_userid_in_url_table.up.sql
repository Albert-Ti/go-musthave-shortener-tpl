ALTER TABLE shorten_url ADD COLUMN user_id VARCHAR(255);

CREATE INDEX idx_shorten_url_user_id ON shorten_url(user_id);