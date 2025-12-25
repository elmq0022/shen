BEGIN;

CREATE TABLE IF NOT EXISTS shen_user(
    id serial PRIMARY KEY,
    username VARCHAR (128) UNIQUE NOT NULL,
    hashed_password VARCHAR (255) NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    role INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_role FOREIGN KEY (role) REFERENCES shen_user_role(id) ON DELETE RESTRICT
);

CREATE INDEX idx_user_role ON shen_user(role);

-- Create trigger for shen_user
CREATE TRIGGER update_shen_user_updated_at
    BEFORE UPDATE ON shen_user
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
