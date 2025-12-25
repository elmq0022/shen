BEGIN;

CREATE TABLE IF NOT EXISTS shen_group (
    id serial PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_name ON shen_group(name);

-- Create trigger for shen_group
CREATE TRIGGER update_shen_group_updated_at
    BEFORE UPDATE ON shen_group
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
