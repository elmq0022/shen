-- name: GetGroupByID :one
SELECT * FROM shen_group
WHERE id = $1 LIMIT 1;

-- name: GetGroupByName :one
SELECT * FROM shen_group
WHERE name = $1 LIMIT 1;

-- name: ListGroups :many
SELECT * FROM shen_group
ORDER BY name;

-- name: ListActiveGroups :many
SELECT * FROM shen_group
WHERE active = true
ORDER BY name;

-- name: CreateGroup :one
INSERT INTO shen_group (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: UpdateGroup :exec
UPDATE shen_group
  set name = $2,
  active = $3
WHERE id = $1;

-- name: DeactivateGroup :exec
UPDATE shen_group
  set active = false
WHERE id = $1;

-- name: DeleteGroup :exec
DELETE FROM shen_group
WHERE id = $1;
