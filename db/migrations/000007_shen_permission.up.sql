BEGIN;

-- Application roles define what users can do within an application
-- These are separate from shen_user_role which controls access to Shen itself
CREATE TABLE IF NOT EXISTS shen_application_role (
    id serial PRIMARY KEY,
    priority INTEGER UNIQUE NOT NULL,
    name VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_application_role_name_lowercase CHECK (name = LOWER(name))
);

CREATE INDEX shen_application_role_priority_idx ON shen_application_role(priority);

-- Create trigger for shen_application_role
CREATE TRIGGER update_shen_application_role_updated_at
    BEFORE UPDATE ON shen_application_role
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default application roles
INSERT INTO shen_application_role (priority, name) VALUES
    (100, 'authenticated'),
    (200, 'viewer'),
    (300, 'auditor'),
    (400, 'operator'),
    (500, 'admin');

COMMIT;
