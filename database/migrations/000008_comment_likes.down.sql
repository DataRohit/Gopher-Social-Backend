DROP INDEX IF EXISTS idx_comment_likes_user_id;

DROP INDEX IF EXISTS idx_comment_likes_comment_id;

DROP TABLE IF EXISTS comment_likes;

DROP EXTENSION IF EXISTS "uuid-ossp";