-- name: ListAllGroupApplicationRoles :many
SELECT
    gar.id,
    g.name AS group_name,
    a.name AS application_name,
    r.name AS role_name,
    gar.created_at,
    gar.updated_at
FROM
    shen_group_application_role gar
    JOIN shen_application a ON gar.application_id = a.id
    JOIN shen_group g ON gar.group_id = g.id
    JOIN shen_application_role r ON gar.role_id = r.id
WHERE
    (sqlc.arg(cursor_group_name)::text IS NULL OR
     g.name > sqlc.arg(cursor_group_name) OR
     (g.name = sqlc.arg(cursor_group_name) AND a.name > sqlc.arg(cursor_application_name)) OR
     (g.name = sqlc.arg(cursor_group_name) AND a.name = sqlc.arg(cursor_application_name) AND r.name > sqlc.arg(cursor_role_name)))
ORDER BY
    g.name,
    a.name,
    r.name
LIMIT $1;

-- name: ListGroupApplicationRolesByGroup :many
SELECT
    gar.id,
    a.name AS application_name,
    r.name AS role_name,
    gar.created_at,
    gar.updated_at
FROM
    shen_group_application_role gar
    JOIN shen_application a ON gar.application_id = a.id
    JOIN shen_application_role r ON gar.role_id = r.id
WHERE
    gar.group_id = sqlc.arg(group_id)
    AND (sqlc.arg(cursor_application_name)::text IS NULL OR
         a.name > sqlc.arg(cursor_application_name) OR
         (a.name = sqlc.arg(cursor_application_name) AND r.name > sqlc.arg(cursor_role_name)))
ORDER BY
    a.name,
    r.name
LIMIT $1;

-- name: ListGroupApplicationRolesByApplication :many
SELECT
    gar.id,
    g.name AS group_name,
    r.name AS role_name,
    gar.created_at,
    gar.updated_at
FROM
    shen_group_application_role gar
    JOIN shen_group g ON gar.group_id = g.id
    JOIN shen_application_role r ON gar.role_id = r.id
WHERE
    gar.application_id = sqlc.arg(application_id)
    AND (sqlc.arg(cursor_group_name)::text IS NULL OR
         g.name > sqlc.arg(cursor_group_name) OR
         (g.name = sqlc.arg(cursor_group_name) AND r.name > sqlc.arg(cursor_role_name)))
ORDER BY
    g.name,
    r.name
LIMIT $1;

-- name: CountGroupApplicationRolesByGroup :one
SELECT
    COUNT(*)
FROM
    shen_group_application_role
WHERE
    group_id = $1;

-- name: CountGroupApplicationRolesByApplication :one
SELECT
    COUNT(*)
FROM
    shen_group_application_role
WHERE
    application_id = $1;

-- name: CountAllGroupApplicationRoles :one
SELECT
    COUNT(*)
FROM
    shen_group_application_role;

-- name: GetGroupApplicationRoleByID :one
SELECT
    id,
    group_id,
    application_id,
    role_id,
    created_at,
    updated_at
FROM
    shen_group_application_role
WHERE
    id = $1
LIMIT 1;

-- name: GetGroupApplicationRole :one
SELECT
    id,
    group_id,
    application_id,
    role_id,
    created_at,
    updated_at
FROM
    shen_group_application_role
WHERE
    group_id = $1
    AND application_id = $2
    AND role_id = $3
LIMIT 1;

-- name: AddGroupApplicationRole :one
INSERT INTO shen_group_application_role(group_id, application_id, role_id)
    VALUES ($1, $2, $3)
RETURNING
    id,
    group_id,
    application_id,
    role_id,
    created_at,
    updated_at;

-- name: DeleteGroupApplicationRole :exec
DELETE FROM shen_group_application_role
WHERE group_id = $1
    AND application_id = $2
    AND role_id = $3;

-- name: DeleteAllGroupApplicationRoles :exec
DELETE FROM shen_group_application_role
WHERE group_id = $1
    AND application_id = $2;

-- name: GetUserApplicationRoles :many
SELECT DISTINCT
    r.id,
    r.priority,
    r.name,
    r.created_at,
    r.updated_at
FROM
    shen_group_application_role gar
    JOIN shen_user_group_member ugm ON gar.group_id = ugm.group_id
    JOIN shen_application_role r ON gar.role_id = r.id
WHERE
    ugm.user_id = $1
    AND gar.application_id = $2
ORDER BY
    r.priority DESC;

-- name: GetUserGroups :many
SELECT DISTINCT
    g.id,
    g.name,
    g.active,
    g.created_at,
    g.updated_at
FROM
    shen_user_group_member ugm
    JOIN shen_group g ON ugm.group_id = g.id
WHERE
    ugm.user_id = $1
    AND g.active = true
ORDER BY
    g.name;
