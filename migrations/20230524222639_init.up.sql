create table users
(
    telegram_id    bigint not null
        constraint users_pk
            primary key,
    fastmail_token text
);
