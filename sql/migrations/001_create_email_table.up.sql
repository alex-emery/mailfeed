PRAGMA foreign_keys = ON;

create table feed (
    id text primary key,
    name text not null
);

CREATE TABLE email (
    id INTEGER PRIMARY KEY,
    date text NOT NULL,
    recipient text NOT NULL,
    sender text NOT NULL,
    subject text NOT NULL,
    description text NOT NULL
);

create table feed_item (
    id text primary key,
    name text not null,
    feed_id text not null references feed(id) ON DELETE CASCADE,
    subject text not null,
    body text not null,
    date text not null
);