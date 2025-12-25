BEGIN;

CREATE TABLE IF NOT EXISTS shen_application (
    id serial PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX shen_application_name_idx ON shen_application(name);

-- Create trigger for shen_application
CREATE TRIGGER update_shen_application_updated_at
    BEFORE UPDATE ON shen_application
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
