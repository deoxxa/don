create table users (
  id text not null primary key,
  created_at datetime not null,
  username text not null unique,
  email text not null unique,
  hash text not null,
  display_name text,
  avatar text,
  summary text
);

create table people (
  id text not null primary key,
  host text not null,
  first_seen datetime not null,
  permalink text not null,
  display_name text,
  avatar text,
  summary text
);

create table objects (
  id text not null primary key,
  name text,
  summary text,
  representative_image text,
  permalink text,
  object_type text,
  content text
);

create table activities (
  id text not null primary key,
  permalink text not null,
  actor references people (id),
  object not null references objects (id),
  verb text not null,
  time datetime not null,
  title text,
  in_reply_to_id text,
  in_reply_to_url text
);

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

create table pubsub_documents (
  created_at datetime not null,
  xml text not null
);
