-- name: GetEmail :one
SELECT
    *
FROM
    email
WHERE
    id = ?
LIMIT
    1;

-- name: ListEmails :many
SELECT
    *
FROM
    email
ORDER BY
    date;

-- name: CreateEmail :one
INSERT INTO
    email (
        id,
        date,
        recipient,
        sender,
        subject,
        description
    )
VALUES
    (?, ?, ?, ?, ?, ?) RETURNING *;

