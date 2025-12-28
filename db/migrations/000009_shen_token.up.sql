BEGIN;

CREATE TABLE IF NOT EXISTS shen_token (
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hashed_token VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    application_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMP,
    CONSTRAINT fk_token_user_id FOREIGN KEY (user_id) REFERENCES shen_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_token_application_id FOREIGN KEY (application_id) REFERENCES shen_application(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_application_name UNIQUE (user_id, application_id, name),
    CONSTRAINT unique_hashed_token UNIQUE (hashed_token)
);

CREATE INDEX shen_token_user_id_idx ON shen_token(user_id);
CREATE INDEX shen_token_application_id_idx ON shen_token(application_id);
CREATE INDEX shen_token_name_idx ON shen_token(name);
CREATE INDEX shen_token_hashed_token_idx ON shen_token(hashed_token);
CREATE INDEX shen_token_created_at_idx ON shen_token(created_at);
CREATE INDEX shen_token_expires_at_idx ON shen_token(expires_at);
CREATE INDEX shen_token_revoked_idx ON shen_token(revoked);

COMMIT;
