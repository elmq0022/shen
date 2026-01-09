-- name: GetGroupByID :one
SELECT
  id,
  name,
  active,
  created_at,
  updated_at
FROM
  shen_group
WHERE
  id = $1
LIMIT 1;

-- name: GetGroupByName :one
SELECT
  id,
  name,
  active,
  created_at,
  updated_at
FROM
  shen_group
WHERE
  name = $1
LIMIT 1;

-- name: ListGroups :many
SELECT
  id,
  name,
  active,
  created_at,
  updated_at
FROM
  shen_group
WHERE
  ($1::text = '' OR name > $1)
ORDER BY
  name
LIMIT $2;

-- name: ListActiveGroups :many
SELECT
  id,
  name,
  active,
  created_at,
  updated_at
FROM
  shen_group
WHERE
  active = TRUE
  AND ($1::text = '' OR name > $1)
ORDER BY
  name
LIMIT $2;

-- name: CreateGroup :one
INSERT INTO shen_group(name)
  VALUES ($1)
RETURNING
  id, name, active, created_at, updated_at;

-- name: UpdateGroup :exec
UPDATE
  shen_group
SET
  name = $2,
  active = $3
WHERE
  id = $1;

-- name: DeactivateGroup :exec
UPDATE
  shen_group
SET
  active = FALSE
WHERE
  id = $1;

-- name: DeleteGroup :exec
DELETE FROM shen_group
WHERE id = $1;

