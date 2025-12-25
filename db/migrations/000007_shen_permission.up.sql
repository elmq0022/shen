BEGIN;

CREATE TABLE IF NOT EXISTS shen_permission (
    id serial PRIMARY KEY,
    priority INTEGER UNIQUE NOT NULL,
    name VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX shen_permission_priority_idx ON shen_permission(priority);

-- Create trigger for shen_permission
CREATE TRIGGER update_shen_permission_updated_at
    BEFORE UPDATE ON shen_permission
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default permissions
INSERT INTO shen_permission (priority, name) VALUES
    (100, 'authenticated'),
    (200, 'viewer'),
    (300, 'auditor'),
    (400, 'operator'),
    (500, 'admin');

COMMIT;
