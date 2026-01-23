-- name: GetTokenByID :one
SELECT
    id,
    name,
    hashed_token,
    user_id,
    application_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_token
WHERE
    id = $1
LIMIT 1;

-- name: GetTokenByHashedToken :one
SELECT
    id,
    name,
    hashed_token,
    user_id,
    application_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_token
WHERE
    hashed_token = $1
LIMIT 1;

-- name: GetTokenByUserApplicationName :one
SELECT
    id,
    name,
    hashed_token,
    user_id,
    application_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_token
WHERE
    user_id = $1
    AND application_id = $2
    AND name = $3
LIMIT 1;

-- name: ListTokensByUser :many
SELECT
    t.id,
    t.name,
    t.user_id,
    a.name AS application_name,
    t.created_at,
    t.expires_at,
    t.revoked,
    t.revoked_at
FROM
    shen_token t
JOIN
    shen_application a ON t.application_id = a.id
WHERE
    t.user_id = sqlc.arg(user_id)
    AND (sqlc.arg(cursor_id) = 0 OR t.id > sqlc.arg(cursor_id))
ORDER BY
    t.id ASC
LIMIT $1;

-- name: ListTokensByApplication :many
SELECT
    t.id,
    t.name,
    t.user_id,
    a.name AS application_name,
    t.created_at,
    t.expires_at,
    t.revoked,
    t.revoked_at
FROM
    shen_token t
JOIN
    shen_application a ON t.application_id = a.id
WHERE
    t.application_id = sqlc.arg(application_id)
    AND (sqlc.arg(cursor_id) = 0 OR t.id > sqlc.arg(cursor_id))
ORDER BY
    t.id ASC
LIMIT $1;

-- name: ListTokensByUserApplication :many
SELECT
    t.id,
    t.name,
    t.user_id,
    a.name AS application_name,
    t.created_at,
    t.expires_at,
    t.revoked,
    t.revoked_at
FROM
    shen_token t
JOIN
    shen_application a ON t.application_id = a.id
WHERE
    t.user_id = sqlc.arg(user_id)
    AND t.application_id = sqlc.arg(application_id)
    AND (sqlc.arg(cursor_id) = 0 OR t.id > sqlc.arg(cursor_id))
ORDER BY
    t.id ASC
LIMIT $1;

-- name: ListActiveTokensByUser :many
SELECT
    t.id,
    t.name,
    t.user_id,
    a.name AS application_name,
    t.created_at,
    t.expires_at,
    t.revoked,
    t.revoked_at
FROM
    shen_token t
JOIN
    shen_application a ON t.application_id = a.id
WHERE
    t.user_id = sqlc.arg(user_id)
    AND t.revoked = FALSE
    AND t.expires_at > NOW()
    AND (sqlc.arg(cursor_id) = 0 OR t.id > sqlc.arg(cursor_id))
ORDER BY
    t.id ASC
LIMIT $1;

-- name: ListActiveTokensByUserApplication :many
SELECT
    t.id,
    t.name,
    t.user_id,
    a.name AS application_name,
    t.created_at,
    t.expires_at,
    t.revoked,
    t.revoked_at
FROM
    shen_token t
JOIN
    shen_application a ON t.application_id = a.id
WHERE
    t.user_id = sqlc.arg(user_id)
    AND t.application_id = sqlc.arg(application_id)
    AND t.revoked = FALSE
    AND t.expires_at > NOW()
    AND (sqlc.arg(cursor_id) = 0 OR t.id > sqlc.arg(cursor_id))
ORDER BY
    t.id ASC
LIMIT $1;

-- name: CountTokensByUser :one
SELECT
    COUNT(*)
FROM
    shen_token
WHERE
    user_id = $1;

-- name: CountTokensByApplication :one
SELECT
    COUNT(*)
FROM
    shen_token
WHERE
    application_id = $1;

-- name: CountTokensByUserApplication :one
SELECT
    COUNT(*)
FROM
    shen_token
WHERE
    user_id = $1
    AND application_id = $2;

-- name: CountActiveTokensByUser :one
SELECT
    COUNT(*)
FROM
    shen_token
WHERE
    user_id = $1
    AND revoked = FALSE
    AND expires_at > NOW();

-- name: CreateToken :one
INSERT INTO shen_token(name, hashed_token, user_id, application_id, expires_at)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    id, name, hashed_token, user_id, application_id, created_at, expires_at, revoked, revoked_at;

-- name: RevokeToken :exec
UPDATE
    shen_token
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    id = $1;

-- name: RevokeTokenByHashedToken :exec
UPDATE
    shen_token
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    hashed_token = $1;

-- name: RevokeAllUserTokens :exec
UPDATE
    shen_token
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    user_id = $1
    AND revoked = FALSE;

-- name: RevokeAllUserApplicationTokens :exec
UPDATE
    shen_token
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    user_id = $1
    AND application_id = $2
    AND revoked = FALSE;

-- name: DeleteToken :exec
DELETE FROM shen_token
WHERE id = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM shen_token
WHERE expires_at < NOW();

-- name: DeleteRevokedTokens :exec
DELETE FROM shen_token
WHERE revoked = TRUE;

-- name: IsTokenValid :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            shen_token
        WHERE
            hashed_token = $1
            AND revoked = FALSE
            AND expires_at > NOW());

