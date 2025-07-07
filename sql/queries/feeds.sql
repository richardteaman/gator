-- name: CreateFeed :one
insert into feeds (
    id,
    created_at,
    updated_at,
    name,
    url,
    user_id
) values ( $1,$2,$3,$4,$5,$6)
returning *;


-- name: GetFeeds :many
select * from feeds;

-- name: GetFeedsWithUsers :many
select 
    f.id as feed_id,
    f.name as feed_name,
    f.url as feed_url,
    u.id as user_id,
    u.name as user_name,
    f.last_fetched_at as last_fetched_at
from feeds f
join users u on f.user_id = u.id;

-- name: GetFeedByURL :one
select * from feeds
where url = $1
limit 1;

-- name: MarkFeedFetched :exec
update feeds 
set last_fetched_at = now(), updated_at = now()
where id = $1;

-- name: GetNextFeedToFetch :one
select * from feeds
order by last_fetched_at nulls first,updated_at asc
limit 1;