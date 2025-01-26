DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;

DROP FUNCTION IF EXISTS update_updated_at_column;

DROP INDEX IF EXISTS idx_profiles_user_id;

DROP INDEX IF EXISTS idx_profiles_id;

DROP TABLE IF EXISTS profiles;

DROP EXTENSION IF EXISTS "uuid-ossp";