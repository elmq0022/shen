BEGIN;

CREATE TABLE IF NOT EXISTS shen_jwt_keys (
    id SERIAL PRIMARY KEY,
    kid TEXT NOT NULL UNIQUE,
    encrypted_private_key BYTEA NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active_for_signing BOOLEAN NOT NULL DEFAULT false,
    active_for_verification BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_shen_jwt_keys_active_for_signing
    ON shen_jwt_keys(active_for_signing)
    WHERE active_for_signing = true;

CREATE INDEX idx_shen_jwt_keys_active_for_verification
    ON shen_jwt_keys(active_for_verification)
    WHERE active_for_verification = true;

COMMIT;
