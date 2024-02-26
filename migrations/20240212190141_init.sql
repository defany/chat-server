-- +goose Up
-- +goose StatementBegin
create table if not exists chats(
    id bigserial primary key,
    title text not null
);

create table if not exists users_chats(
    chat_id bigserial references chats(id),
    user_id numeric(12, 0) constraint positive_users_chats_user_id check ( user_id > 0 ),

    primary key (chat_id, user_id)
);

create table if not exists chats_messages(
    id bigserial primary key,
    chat_id bigserial references chats(id),
    user_id numeric(12, 0) constraint positive_users_chats_user_id check ( user_id > 0 ),
    text text not null,
    timestamp timestamp not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists chats_messages;
drop table if exists users_chats;
drop table if exists chats;
-- +goose StatementEnd
