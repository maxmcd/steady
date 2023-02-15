-- name: CreateUser :one
INSERT INTO users (email, username)
VALUES (?, ?)
RETURNING *;
-- name: GetUser :one
SELECT *
FROM users
WHERE id = ?;
-- name: GetUserByEmailOrUsername :one
SELECT *
FROM users
WHERE username = ?
    OR email = ?;
--
-- name: CreateLoginToken :one
INSERT INTO login_tokens (user_id, token)
values (?, ?)
RETURNING *;
-- name: GetLoginToken :one
SELECT *
FROM login_tokens
WHERE token = ?;
-- name: DeleteLoginToken :exec
DELETE FROM login_tokens
where token = ?;
--
-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, token)
values (?, ?)
RETURNING *;
-- name: GetUserSession :one
SELECT *
FROM user_sessions
WHERE token = ?;
-- name: DeleteUserSession :exec
DELETE FROM user_sessions
where token = ?;
--
-- name: GetUserApplications :many
SELECT *
FROM applications
WHERE user_id = ?;
-- name: GetApplication :one
SELECT *
FROM applications
WHERE name = ?;
-- name: CreateApplication :one
INSERT into applications (name, user_id)
values (?, ?)
RETURNING *;
