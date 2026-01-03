BEGIN;

-- Maps Shen groups to application roles
-- A group can have multiple roles for an application (e.g., both viewer and auditor)
CREATE TABLE IF NOT EXISTS shen_group_application_role (
    id serial PRIMARY KEY,
    group_id INTEGER NOT NULL,
    application_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_gar_group_id FOREIGN KEY (group_id) REFERENCES shen_group(id) ON DELETE CASCADE,
    CONSTRAINT fk_gar_application_id FOREIGN KEY (application_id) REFERENCES shen_application(id) ON DELETE CASCADE,
    CONSTRAINT fk_gar_role_id FOREIGN KEY (role_id) REFERENCES shen_application_role(id) ON DELETE RESTRICT,
    CONSTRAINT unique_group_application_role UNIQUE (group_id, application_id, role_id)
);

CREATE INDEX shen_group_application_role_group_id_idx ON shen_group_application_role(group_id);
CREATE INDEX shen_group_application_role_application_id_idx ON shen_group_application_role(application_id);
CREATE INDEX shen_group_application_role_role_id_idx ON shen_group_application_role(role_id);

-- Create trigger for shen_group_application_role
CREATE TRIGGER update_shen_group_application_role_updated_at
    BEFORE UPDATE ON shen_group_application_role
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;