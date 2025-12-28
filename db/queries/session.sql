-- name: GetSessionByID :one
SELECT
    id,
    hashed_token,
    user_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_session
WHERE
    id = $1
LIMIT 1;

-- name: GetSessionByHashedToken :one
SELECT
    id,
    hashed_token,
    user_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_session
WHERE
    hashed_token = $1
LIMIT 1;

-- name: ListSessionsByUser :many
SELECT
    id,
    hashed_token,
    user_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_session
WHERE
    user_id = $1
ORDER BY
    created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveSessions :many
SELECT
    id,
    hashed_token,
    user_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_session
WHERE
    revoked = FALSE
    AND expires_at > NOW()
ORDER BY
    created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveSessionsByUser :many
SELECT
    id,
    hashed_token,
    user_id,
    created_at,
    expires_at,
    revoked,
    revoked_at
FROM
    shen_session
WHERE
    user_id = $1
    AND revoked = FALSE
    AND expires_at > NOW()
ORDER BY
    created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSessionsByUser :one
SELECT
    COUNT(*)
FROM
    shen_session
WHERE
    user_id = $1;

-- name: CountActiveSessions :one
SELECT
    COUNT(*)
FROM
    shen_session
WHERE
    revoked = FALSE
    AND expires_at > NOW();

-- name: CountActiveSessionsByUser :one
SELECT
    COUNT(*)
FROM
    shen_session
WHERE
    user_id = $1
    AND revoked = FALSE
    AND expires_at > NOW();

-- name: CreateSession :one
INSERT INTO shen_session(hashed_token, user_id, expires_at)
    VALUES ($1, $2, $3)
RETURNING
    id, hashed_token, user_id, created_at, expires_at, revoked, revoked_at;

-- name: RevokeSession :exec
UPDATE
    shen_session
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    id = $1;

-- name: RevokeSessionByHashedToken :exec
UPDATE
    shen_session
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    hashed_token = $1;

-- name: RevokeAllUserSessions :exec
UPDATE
    shen_session
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE
    user_id = $1
    AND revoked = FALSE;

-- name: DeleteSession :exec
DELETE FROM shen_session
WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM shen_session
WHERE expires_at < NOW();

-- name: DeleteRevokedSessions :exec
DELETE FROM shen_session
WHERE revoked = TRUE;

-- name: IsSessionValid :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            shen_session
        WHERE
            hashed_token = $1
            AND revoked = FALSE
            AND expires_at > NOW());

