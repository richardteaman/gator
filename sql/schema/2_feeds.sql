-- +goose Up
create table feeds (
    id UUID primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    name TEXT not null ,
    url text not null unique,
    user_id UUID not null references users(id) on delete cascade
);

-- +goose Down
drop table feeds;

