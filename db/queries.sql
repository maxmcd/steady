-- name: CreateUser :one
INSERT INTO users (email, username)
VALUES (?, ?)
RETURNING *;
-- name: GetUser :one
SELECT *
FROM users
WHERE id = ?;
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
-- name: GetUserByEmailOrUsername :one
SELECT *
FROM users
WHERE username = ?
    OR email = ?;
-- name: CreateService :one
INSERT INTO services (name, user_id)
VALUES (?, ?)
RETURNING *;
-- name: GetService :one
SELECT *
FROM services
WHERE user_id = ?
    and id = ?;
-- name: CreateServiceVersion :one
INSERT INTO service_versions (service_id, version, source)
VALUES (?, ?, ?)
RETURNING *;
-- name: GetUserServices :many
SELECT *
FROM services
WHERE user_id = ?;
-- name: GetServiceVersions :many
SELECT *
FROM service_versions
WHERE service_id = ?;
-- name: GetServiceVersion :one
SELECT *
FROM service_versions
WHERE id = ?;
-- name: GetUserApplications :many
SELECT *
FROM applications
WHERE user_id = ?;
