CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE comment_likes (
    user_id UUID NOT NULL,
    comment_id UUID NOT NULL,
    liked BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, comment_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX idx_comment_likes_user_id ON comment_likes (user_id);
CREATE INDEX idx_comment_likes_comment_id ON comment_likes (comment_id);
