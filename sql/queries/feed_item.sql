-- name: CreateFeedItem :one
INSERT into
    feed_item(
        id, 
        name,
        feed_id,
        subject,
        body,
        date
        )
VALUES
    (?, ?, ?,?,?,?) RETURNING *;

-- name: GetFeedItem :one
SELECT
    *
FROM
    feed_item 
where
    id= ?
limit
    1;

-- name: ListFeedItems :many
SELECT
    *
FROM
    feed_item
WHERE
    feed_id = ?;

