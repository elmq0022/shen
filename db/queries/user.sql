-- name: GetUserByID :one
SELECT id, username, hashed_password, active, role, created_at, updated_at FROM shen_user
WHERE id = $1 LIMIT 1;

-- name: GetUserByUserName :one
SELECT id, username, hashed_password, active, role, created_at, updated_at FROM shen_user
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
SELECT id, username, hashed_password, active, role, created_at, updated_at FROM shen_user
ORDER BY username
LIMIT $1 OFFSET $2;

-- name: ListActiveUsers :many
SELECT id, username, hashed_password, active, role, created_at, updated_at FROM shen_user 
where active = true
ORDER BY username
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO shen_user (
    username,
    hashed_password,
    role
) VALUES (
    $1, $2, $3
)
RETURNING id, username, hashed_password, active, role, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE shen_user
  set hashed_password = $2
WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE shen_user
  set role = $2
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE shen_user
  set active = false
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE shen_user
  set active = true
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM shen_user
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM shen_user;

-- name: CountActiveUsers :one
SELECT COUNT(*) FROM shen_user
WHERE active = true;

-- name: ListUsersByRole :many
SELECT id, username, hashed_password, active, role, created_at, updated_at FROM shen_user
WHERE role = $1
ORDER BY username
LIMIT $2 OFFSET $3;

-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM shen_user WHERE username = $1);
