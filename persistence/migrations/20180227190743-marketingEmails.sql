
-- +migrate Up
create table marketing_emails (
    id serial primary key,
    email text unique,
    created_at_utc timestamp not null,
    updated_at_utc timestamp null
);
-- +migrate Down
drop table marketing_emails;