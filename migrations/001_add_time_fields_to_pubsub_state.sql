alter table pubsub_state rename to pubsub_state_old;

create table pubsub_state (
  id text not null primary key,
  hub text not null,
  topic text not null,
  callback_url text not null,
  created_at datetime not null,
  updated_at datetime not null,
  expires_at datetime,
  unique (hub, topic)
);

insert into pubsub_state
  (id, hub, topic, callback_url, created_at, updated_at, expires_at)
  select id, hub, topic, callback_url, 0, 0, expires_at from pubsub_state_old;

drop table pubsub_state_old;
