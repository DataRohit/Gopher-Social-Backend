DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS roles;
DROP EXTENSION IF EXISTS "uuid-ossp";