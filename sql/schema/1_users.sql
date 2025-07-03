--  +goose Up
Create table users (
    id UUID Primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    name TEXT unique not null

);



-- +goose Down
drop table users;