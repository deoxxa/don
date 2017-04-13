create table users (
  id text not null primary key,
  created_at datetime not null,
  username text not null unique,
  email text not null unique,
  hash text not null,
  display_name text,
  avatar text
);
