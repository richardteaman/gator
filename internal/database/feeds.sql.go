// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: feeds.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createFeed = `-- name: CreateFeed :one
insert into feeds (
    id,
    created_at,
    updated_at,
    name,
    url,
    user_id
) values ( $1,$2,$3,$4,$5,$6)
returning id, created_at, updated_at, name, url, user_id, last_fetched_at
`

type CreateFeedParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Url       string
	UserID    uuid.UUID
}

func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
	row := q.db.QueryRowContext(ctx, createFeed,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Url,
		arg.UserID,
	)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const getFeedByURL = `-- name: GetFeedByURL :one
select id, created_at, updated_at, name, url, user_id, last_fetched_at from feeds
where url = $1
limit 1
`

func (q *Queries) GetFeedByURL(ctx context.Context, url string) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getFeedByURL, url)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const getFeeds = `-- name: GetFeeds :many
select id, created_at, updated_at, name, url, user_id, last_fetched_at from feeds
`

func (q *Queries) GetFeeds(ctx context.Context) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getFeeds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
			&i.LastFetchedAt,
		); err != nil {
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

const getFeedsWithUsers = `-- name: GetFeedsWithUsers :many
select 
    f.id as feed_id,
    f.name as feed_name,
    f.url as feed_url,
    u.id as user_id,
    u.name as user_name,
    f.last_fetched_at as last_fetched_at
from feeds f
join users u on f.user_id = u.id
`

type GetFeedsWithUsersRow struct {
	FeedID        uuid.UUID
	FeedName      string
	FeedUrl       string
	UserID        uuid.UUID
	UserName      string
	LastFetchedAt sql.NullTime
}

func (q *Queries) GetFeedsWithUsers(ctx context.Context) ([]GetFeedsWithUsersRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedsWithUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedsWithUsersRow
	for rows.Next() {
		var i GetFeedsWithUsersRow
		if err := rows.Scan(
			&i.FeedID,
			&i.FeedName,
			&i.FeedUrl,
			&i.UserID,
			&i.UserName,
			&i.LastFetchedAt,
		); err != nil {
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

const getNextFeedToFetch = `-- name: GetNextFeedToFetch :one
select id, created_at, updated_at, name, url, user_id, last_fetched_at from feeds
order by last_fetched_at nulls first,updated_at asc
limit 1
`

func (q *Queries) GetNextFeedToFetch(ctx context.Context) (Feed, error) {
	row := q.db.QueryRowContext(ctx, getNextFeedToFetch)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.UserID,
		&i.LastFetchedAt,
	)
	return i, err
}

const markFeedFetched = `-- name: MarkFeedFetched :exec
update feeds 
set last_fetched_at = now(), updated_at = now()
where id = $1
`

func (q *Queries) MarkFeedFetched(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, markFeedFetched, id)
	return err
}
