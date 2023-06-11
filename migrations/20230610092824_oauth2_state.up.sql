create table oauth2_states(
  state text not null,
  code_verifier text not null,
  telegram_id bigint not null references users(telegram_id)
);