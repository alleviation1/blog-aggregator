-- +goose Up
CREATE TABLE posts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMP NOT NULL,
    updated_at   TIMESTAMP NOT NULL,
    title        TEXT NOT NULL UNIQUE,
    url          TEXT NOT NULL,
    description  TEXT,
    published_at TIMESTAMP,
    feed_id      UUID NOT NULL,

    CONSTRAINT fk_feed_id_posts
     FOREIGN KEY (feed_id)
     REFERENCES feeds(id)
     ON UPDATE CASCADE
     ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;