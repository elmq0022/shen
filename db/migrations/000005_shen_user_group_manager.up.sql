BEGIN;

CREATE TABLE IF NOT EXISTS shen_user_group_manager(
    id serial PRIMARY KEY,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_ugmgr_user_id FOREIGN KEY (user_id) REFERENCES shen_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_ugmgr_group_id FOREIGN KEY (group_id) REFERENCES shen_group(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_group_manager UNIQUE (user_id, group_id) 
);

CREATE INDEX idx_user_id_manger_user ON shen_user_group_manager(user_id);
CREATE INDEX idx_group_id_manager_group ON shen_user_group_manager(group_id);

-- Create trigger for shen_user_group_manager
CREATE TRIGGER update_shen_user_group_manager_updated_at
    BEFORE UPDATE ON shen_user_group_manager
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
