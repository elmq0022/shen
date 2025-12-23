BEGIN;

CREATE TABLE IF NOT EXISTS shen_group_application_permission (
    id serial PRIMARY KEY,
    group_id INTEGER NOT NULL,
    application_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_group_id_group_group_application_permission FOREIGN KEY (group_id) REFERENCES shen_group(id) ON DELETE CASCADE,
    CONSTRAINT fk_application_id_group_group_application_permission FOREIGN KEY (application_id) REFERENCES shen_application(id) ON DELETE CASCADE,
    CONSTRAINT fk_permission_id_group_group_application_permission FOREIGN KEY (permission_id) REFERENCES shen_permission(id) ON DELETE RESTRICT,
    CONSTRAINT unique_group_application UNIQUE (group_id, application_id)
);

CREATE INDEX group_id_idx ON shen_group_application_permission(group_id);
CREATE INDEX application_id_idx ON shen_group_application_permission(application_id);
CREATE INDEX permission_id_idx on shen_group_application_permission(permission_id);

COMMIT;