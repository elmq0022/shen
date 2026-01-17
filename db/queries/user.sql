-- name: GetUserByID :one
SELECT
  id,
  username,
  hashed_password,
  active,
  ROLE,
  created_at,
  updated_at
FROM
  shen_user
WHERE
  id = $1
LIMIT 1;

-- name: GetUserByUsername :one
SELECT
  id,
  username,
  hashed_password,
  active,
  ROLE,
  created_at,
  updated_at
FROM
  shen_user
WHERE
  username = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
  id,
  username,
  active,
  ROLE,
  created_at,
  updated_at
FROM
  shen_user
WHERE
  ($1::text = '' OR username > $1)
ORDER BY
  username
LIMIT $2;

-- name: ListActiveUsers :many
SELECT
  su.id,
  su.username,
  su.active,
  sur.name AS role,
  su.created_at,
  su.updated_at
FROM
  shen_user su
JOIN
  shen_user_role sur ON su.role = sur.id
WHERE
  su.active = TRUE
  AND ($1::text = '' OR su.username > $1)
ORDER BY
  su.username
LIMIT $2;

-- name: CreateUser :one
INSERT INTO shen_user(username, hashed_password, role)
  VALUES ($1, $2, $3)
RETURNING
  id, username, active, role, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE
  shen_user
SET
  hashed_password = $2
WHERE
  id = $1;

-- name: UpdateUserRole :exec
UPDATE
  shen_user
SET
  ROLE = $2
WHERE
  id = $1;

-- name: DeactivateUser :exec
UPDATE
  shen_user
SET
  active = FALSE
WHERE
  id = $1;

-- name: ActivateUser :exec
UPDATE
  shen_user
SET
  active = TRUE
WHERE
  id = $1;

-- name: DeleteUser :exec
DELETE FROM shen_user
WHERE id = $1;

-- name: CountUsers :one
SELECT
  COUNT(*)
FROM
  shen_user;

-- name: CountActiveUsers :one
SELECT
  COUNT(*)
FROM
  shen_user
WHERE
  active = TRUE;

-- name: ListUsersByRole :many
SELECT
  id,
  username,
  active,
  ROLE,
  created_at,
  updated_at
FROM
  shen_user
WHERE
  ROLE = $1
  AND ($2::text = '' OR username > $2)
ORDER BY
  username
LIMIT $3;

-- name: CheckUsernameExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      shen_user
    WHERE
      username = $1);

