-- name: GetRoleByID :one
SELECT
    id,
    name,
    created_at,
    updated_at
FROM
    shen_user_role
WHERE
    id = $1
LIMIT 1;

-- name: GetRoleByName :one
SELECT
    id,
    name,
    created_at,
    updated_at
FROM
    shen_user_role
WHERE
    name = $1
LIMIT 1;

-- name: ListRoles :many
SELECT
    id,
    name,
    created_at,
    updated_at
FROM
    shen_user_role
ORDER BY
    name;
