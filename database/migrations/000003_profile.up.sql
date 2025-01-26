CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    website TEXT,
    github TEXT,
    linkedin TEXT,
    twitter TEXT,
    google_scholar TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_profiles_id ON profiles (id);
CREATE INDEX idx_profiles_user_id ON profiles (user_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_profiles_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();
