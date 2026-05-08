CREATE TABLE shorten_url (
  id SERIAL PRIMARY KEY,
  key VARCHAR(100) UNIQUE NOT NULL,
  url VARCHAR(255)  NOT NULL,
  correlation_id VARCHAR(255)
);

CREATE UNIQUE INDEX idx_shorten_url ON shorten_url(url);