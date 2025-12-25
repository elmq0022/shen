BEGIN;

-- Create the trigger function for updating updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS shen_user_role (
    id serial PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger for shen_user_role
CREATE TRIGGER update_shen_user_role_updated_at
    BEFORE UPDATE ON shen_user_role
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default user roles
INSERT INTO shen_user_role (name) VALUES
    ('service'),
    ('user'),
    ('admin');

COMMIT;
