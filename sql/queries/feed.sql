-- name: CreateFeed :one
INSERT into
    feed (id, name)
VALUES
    (?, ?) RETURNING *;

-- name: GetFeed :one
SELECT
    *
FROM
    feed 
where
    id= ?
limit
    1;

-- name: ListFeeds :many
SELECT
    *
FROM
    feed;