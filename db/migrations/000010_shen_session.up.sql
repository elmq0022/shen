BEGIN;

CREATE TABLE IF NOT EXISTS shen_session (
    id serial PRIMARY KEY,
    hashed_token VARCHAR(64) NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMPTZ,
    CONSTRAINT fk_session_user_id FOREIGN KEY (user_id) REFERENCES shen_user(id) ON DELETE CASCADE,
    CONSTRAINT unique_session_hashed_token UNIQUE (hashed_token)
);

CREATE INDEX shen_session_user_id_idx ON shen_session(user_id);
CREATE INDEX shen_session_hashed_token_idx ON shen_session(hashed_token);
CREATE INDEX shen_session_created_at_idx ON shen_session(created_at);
CREATE INDEX shen_session_expires_at_idx ON shen_session(expires_at);
CREATE INDEX shen_session_revoked_idx ON shen_session(revoked);

COMMIT;
