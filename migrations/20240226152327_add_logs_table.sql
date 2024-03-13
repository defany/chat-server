-- +goose Up
-- +goose StatementBegin
create table if not exists logs(
    action text not null,
    user_id bigserial not null constraint positive_user_id check ( user_id > 0 ),
    timestamp timestamp not null default clock_timestamp()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists logs;
-- +goose StatementEnd
