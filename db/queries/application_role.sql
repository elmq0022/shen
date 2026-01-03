-- name: GetApplicationRoleByID :one
SELECT
    id,
    priority,
    name,
    created_at,
    updated_at
FROM
    shen_application_role
WHERE
    id = $1
LIMIT 1;

-- name: GetApplicationRoleByName :one
SELECT
    id,
    priority,
    name,
    created_at,
    updated_at
FROM
    shen_application_role
WHERE
    name = $1
LIMIT 1;

-- name: ListApplicationRoles :many
SELECT
    id,
    priority,
    name,
    created_at,
    updated_at
FROM
    shen_application_role
ORDER BY
    priority;
