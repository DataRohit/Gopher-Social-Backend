DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;

DROP INDEX IF EXISTS idx_comments_post_id;

DROP INDEX IF EXISTS idx_comments_author_id;

DROP INDEX IF EXISTS idx_comments_id;

DROP TABLE IF EXISTS comments;

DROP EXTENSION IF EXISTS "uuid-ossp";
