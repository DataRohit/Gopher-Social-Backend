DROP INDEX IF EXISTS idx_post_likes_post_id;

DROP INDEX IF EXISTS idx_post_likes_user_id;

DROP TABLE IF EXISTS post_likes;

DROP EXTENSION IF EXISTS "uuid-ossp";
