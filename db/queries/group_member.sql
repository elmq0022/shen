-- name: GetUserGroupMemberByID :one
SELECT
    id,
    user_id,
    group_id,
    created_at,
    updated_at
FROM
    shen_user_group_member
WHERE
    id = $1
LIMIT 1;

-- name: ListGroupsByUser :many
SELECT
    g.id,
    g.name,
    g.active,
    g.created_at,
    g.updated_at
FROM
    shen_user_group_member m
    JOIN shen_group g ON m.group_id = g.id
WHERE
    m.user_id = $1
    AND ($2::text = '' OR g.name > $2)
ORDER BY
    g.name
LIMIT $3;

-- name: ListUsersByGroup :many
SELECT
    u.id,
    u.username,
    u.hashed_password,
    u.active,
    u.role,
    u.created_at,
    u.updated_at
FROM
    shen_user_group_member m
    JOIN shen_user u ON m.user_id = u.id
WHERE
    m.group_id = $1
    AND ($2::text = '' OR u.username > $2)
ORDER BY
    u.username
LIMIT $3;

-- name: ListAllGroupMembers :many
SELECT
    m.id,
    g.name AS group_name,
    u.username AS username,
    m.created_at,
    m.updated_at
FROM
    shen_user_group_member m
    JOIN shen_user u ON m.user_id = u.id
    JOIN shen_group g ON m.group_id = g.id
WHERE
    ($1::text = '' OR g.name > $1 OR (g.name = $1 AND u.username > $2))
ORDER BY
    g.name,
    u.username
LIMIT $3;

-- name: CountGroupsByUser :one
SELECT
    COUNT(*)
FROM
    shen_user_group_member
WHERE
    user_id = $1;

-- name: CountUsersByGroup :one
SELECT
    COUNT(*)
FROM
    shen_user_group_member
WHERE
    group_id = $1;

-- name: CountAllGroupMembers :one
SELECT
    COUNT(*)
FROM
    shen_user_group_member;

-- name: AddUserToGroup :one
INSERT INTO shen_user_group_member(user_id, group_id)
    VALUES ($1, $2)
RETURNING
    id, user_id, group_id, created_at, updated_at;

-- name: RemoveUserFromGroup :exec
DELETE FROM shen_user_group_member
WHERE user_id = $1
    AND group_id = $2;

-- name: IsUserInGroup :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            shen_user_group_member
        WHERE
            user_id = $1
            AND group_id = $2);

