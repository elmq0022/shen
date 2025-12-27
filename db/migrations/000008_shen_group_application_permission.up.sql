BEGIN;

CREATE TABLE IF NOT EXISTS shen_group_application_permission (
    id serial PRIMARY KEY,
    group_id INTEGER NOT NULL,
    application_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_gap_group_id FOREIGN KEY (group_id) REFERENCES shen_group(id) ON DELETE CASCADE,
    CONSTRAINT fk_gap_application_id FOREIGN KEY (application_id) REFERENCES shen_application(id) ON DELETE CASCADE,
    CONSTRAINT fk_gap_permission_id FOREIGN KEY (permission_id) REFERENCES shen_permission(id) ON DELETE RESTRICT,
    CONSTRAINT unique_group_application UNIQUE (group_id, application_id)
);

CREATE INDEX shen_group_application_permission_group_id_idx ON shen_group_application_permission(group_id);
CREATE INDEX shen_group_application_permission_application_id_idx ON shen_group_application_permission(application_id);
CREATE INDEX shen_group_application_permission_permission_id_idx ON shen_group_application_permission(permission_id);

-- Create trigger for shen_group_application_permission
CREATE TRIGGER update_shen_group_application_permission_updated_at
    BEFORE UPDATE ON shen_group_application_permission
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;