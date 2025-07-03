-- name: CreateFeedFollow :one
with inserted as (
    insert into feed_follows (
        id,
        created_at,
        updated_at,
        user_id,
        feed_id
    ) values (
        $1,$2,$3,$4,$5
    )
    returning *
)
select
    inserted.id,
    inserted.user_id,
    inserted.feed_id,
    inserted.created_at,
    inserted.updated_at,
    users.name as user_name,
    feeds.name as feed_name
from inserted
join users on users.id = inserted.user_id
join feeds on feeds.id = inserted.feed_id;


-- name: GetFeedFollowsForUser :many
select  
    feed_follows.id,
    feed_follows.user_id,
    feed_follows.feed_id,
    feed_follows.created_at,
    feed_follows.updated_at,
    users.name as user_name,
    feeds.name as feed_name
from feed_follows
join feeds on feeds.id = feed_follows.feed_id
join users on users.id = feed_follows.user_id
where feed_follows.user_id = $1;

-- name: GetFeedFollowForUserAndFeed :one
SELECT * FROM feed_follows 
WHERE user_id = $1 AND feed_id = $2 
LIMIT 1;