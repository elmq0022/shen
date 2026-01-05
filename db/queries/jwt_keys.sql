-- name: GetJWTKeyByID :one
SELECT
    id,
    kid,
    encrypted_private_key,
    public_key,
    created_at,
    active_for_signing,
    active_for_verification
FROM
    shen_jwt_keys
WHERE
    id = $1
LIMIT 1;

-- name: GetJWTKeyByKID :one
SELECT
    id,
    kid,
    encrypted_private_key,
    public_key,
    created_at,
    active_for_signing,
    active_for_verification
FROM
    shen_jwt_keys
WHERE
    kid = $1
LIMIT 1;

-- name: GetActiveSigningKey :one
SELECT
    id,
    kid,
    encrypted_private_key,
    public_key,
    created_at,
    active_for_signing,
    active_for_verification
FROM
    shen_jwt_keys
WHERE
    active_for_signing = TRUE
LIMIT 1;

-- name: ListJWTKeys :many
SELECT
    id,
    kid,
    encrypted_private_key,
    public_key,
    created_at,
    active_for_signing,
    active_for_verification
FROM
    shen_jwt_keys
ORDER BY
    created_at DESC,
    id DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveVerificationKeys :many
SELECT
    id,
    kid,
    encrypted_private_key,
    public_key,
    created_at,
    active_for_signing,
    active_for_verification
FROM
    shen_jwt_keys
WHERE
    active_for_verification = TRUE
ORDER BY
    created_at DESC,
    id DESC;

-- name: CreateJWTKey :one
INSERT INTO shen_jwt_keys(kid, encrypted_private_key, public_key, active_for_signing, active_for_verification)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    id, kid, encrypted_private_key, public_key, created_at, active_for_signing, active_for_verification;

-- name: DeactivateKeyForSigning :exec
UPDATE
    shen_jwt_keys
SET
    active_for_signing = FALSE
WHERE
    id = $1;

-- name: DeactivateKeyForVerification :exec
UPDATE
    shen_jwt_keys
SET
    active_for_verification = FALSE
WHERE
    id = $1;

-- name: DeactivateAllKeysForSigning :exec
UPDATE
    shen_jwt_keys
SET
    active_for_signing = FALSE
WHERE
    active_for_signing = TRUE;

-- name: DeleteJWTKey :exec
DELETE FROM shen_jwt_keys
WHERE id = $1;

-- name: CountJWTKeys :one
SELECT
    COUNT(*)
FROM
    shen_jwt_keys;
