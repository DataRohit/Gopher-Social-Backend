CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE post_likes (
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    liked BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE INDEX idx_post_likes_user_id ON post_likes (user_id);
CREATE INDEX idx_post_likes_post_id ON post_likes (post_id);
