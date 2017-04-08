create table pubsub_state (
	id text not null primary key,
	hub text not null,
	topic text not null,
	callback_url text not null,
	expires_at datetime,
	unique (hub, topic)
);

create table people (
	feed_url text not null primary key,
	first_seen datetime not null,
	name text,
	display_name text,
	email text,
	summary text,
	note text
);

create table posts (
	feed_url text not null,
	id text not null,
	created_at datetime not null,
	raw_entry text not null,
	primary key (feed_url, id)
);
