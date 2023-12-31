// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: feed.sql

package sqlc

import (
	"context"
)

const createFeed = `-- name: CreateFeed :one
INSERT into
    feed (id, name)
VALUES
    (?, ?) RETURNING id, name
`

type CreateFeedParams struct {
	ID   string
	Name string
}

func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
	row := q.db.QueryRowContext(ctx, createFeed, arg.ID, arg.Name)
	var i Feed
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getFeed = `-- name: GetFeed :one
SELECT
    id, name
FROM
    feed 
where
    id= ?
limit
    1
`

func (q *Queries) GetFeed(ctx context.Context, id string) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getFeed, id)
	var i Feed
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const listFeeds = `-- name: ListFeeds :many
SELECT
    id, name
FROM
    feed
`

func (q *Queries) ListFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, listFeeds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
