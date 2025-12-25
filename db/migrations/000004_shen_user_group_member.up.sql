BEGIN;

CREATE TABLE IF NOT EXISTS shen_user_group_member(
    id serial PRIMARY KEY,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_group_member_user FOREIGN KEY (user_id) REFERENCES shen_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_group_member_group FOREIGN KEY (group_id) REFERENCES shen_group(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_group_member UNIQUE (user_id, group_id)
);

CREATE INDEX idx_user_id_member_user ON shen_user_group_member(user_id);
CREATE INDEX idx_group_id_member_group ON shen_user_group_member(group_id);

-- Create trigger for shen_user_group_member
CREATE TRIGGER update_shen_user_group_member_updated_at
    BEFORE UPDATE ON shen_user_group_member
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
