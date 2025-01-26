DROP INDEX IF EXISTS idx_follows_followee_id;

DROP INDEX IF EXISTS idx_follows_follower_id;

DROP TABLE IF EXISTS follows;

DROP EXTENSION IF EXISTS "uuid-ossp";