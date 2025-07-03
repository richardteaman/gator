-- name: CreateFeed :one
insert into feeds (
    id,
    created_at,
    updated_at,
    name,
    url,
    user_id
) values ( $1,$2,$3,$4,$5,$6)
returning id,created_at,updated_at,name,url,user_id;


-- name: GetFeeds :many
select * from feeds;

-- name: GetFeedsWithUsers :many
select 
    f.id as feed_id,
    f.name as feed_name,
    f.url as feed_url,
    u.id as user_id,
    u.name as user_name
from feeds f
join users u on f.user_id = u.id;

-- name: GetFeedByURL :one
select * from feeds
where url = $1
limit 1;