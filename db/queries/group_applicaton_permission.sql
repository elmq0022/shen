-- name: ListAllGroupApplicationPermissions :many
SELECT
    gap.id,
    g.name AS group_name,
    a.name AS application_name,
    p.name AS permission_name,
    gap.created_at,
    gap.updated_at
FROM
    shen_group_application_permission gap
    JOIN shen_application a ON gap.application_id = a.id
    JOIN shen_group g ON gap.group_id = g.id
    JOIN shen_permission p ON gap.permission_id = p.id
ORDER BY
    g.name,
    a.name,
    p.name
LIMIT $1 OFFSET $2;

-- name: ListGroupApplicationPermissionsByGroup :many
SELECT
    gap.id,
    a.name AS application_name,
    p.name AS permission_name,
    gap.created_at,
    gap.updated_at
FROM
    shen_group_application_permission gap
    JOIN shen_application a ON gap.application_id = a.id
    JOIN shen_permission p ON gap.permission_id = p.id
WHERE
    gap.group_id = $1
ORDER BY
    a.name,
    p.name
LIMIT $2 OFFSET $3;

-- name: ListGroupApplicationPermissionsByApplication :many
SELECT
    gap.id,
    g.name AS group_name,
    p.name AS permission_name,
    gap.created_at,
    gap.updated_at
FROM
    shen_group_application_permission gap
    JOIN shen_group g ON gap.group_id = g.id
    JOIN shen_permission p ON gap.permission_id = p.id
WHERE
    gap.application_id = $1
ORDER BY
    g.name,
    p.name
LIMIT $2 OFFSET $3;

-- name: CountGroupApplicationPermissionsByGroup :one
SELECT
    COUNT(*)
FROM
    shen_group_application_permission
WHERE
    group_id = $1;

-- name: CountGroupApplicationPermissionsByApplication :one
SELECT
    COUNT(*)
FROM
    shen_group_application_permission
WHERE
    application_id = $1;

-- name: CountAllGroupApplicationPermissions :one
SELECT
    COUNT(*)
FROM
    shen_group_application_permission;

-- name: GetGroupApplicationPermissionByID :one
SELECT
    id,
    group_id,
    application_id,
    permission_id,
    created_at,
    updated_at
FROM
    shen_group_application_permission
WHERE
    id = $1
LIMIT 1;

-- name: GetGroupApplicationPermission :one
SELECT
    id,
    group_id,
    application_id,
    permission_id,
    created_at,
    updated_at
FROM
    shen_group_application_permission
WHERE
    group_id = $1
    AND application_id = $2
LIMIT 1;

-- name: SetGroupApplicationPermission :one
INSERT INTO shen_group_application_permission(group_id, application_id, permission_id)
    VALUES ($1, $2, $3)
ON CONFLICT (group_id, application_id)
    DO UPDATE SET
        permission_id = EXCLUDED.permission_id
    RETURNING
        id,
        group_id,
        application_id,
        permission_id,
        created_at,
        updated_at;

-- name: DeleteGroupApplicationPermission :exec
DELETE FROM shen_group_application_permission
WHERE group_id = $1
    AND application_id = $2;

-- name: GetUserApplicationPermission :one
SELECT
    p.id,
    p.priority,
    p.name,
    p.created_at,
    p.updated_at
FROM
    shen_group_application_permission gap
    JOIN shen_user_group_member ugm ON gap.group_id = ugm.group_id
    JOIN shen_permission p ON gap.permission_id = p.id
WHERE
    ugm.user_id = $1
    AND gap.application_id = $2
ORDER BY
    p.priority ASC
LIMIT 1;

