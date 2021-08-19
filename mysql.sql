DROP DATABASE IF EXISTS nakama CASCADE;
CREATE DATABASE IF NOT EXISTS nakama;
SET DATABASE = nakama;

create table  users (
	id serial not null primary key,
    email varchar not null unique,
    username varchar not null unique,
    avatar varchar ,
    followers_count int not null default 0,
    followees_count int not null default 0
);

create table follows (
	follower_id int not null references users,
    followee_id int not null references users,
    primary key ( follower_id , followee_id)
);

create table posts (
	id SERIAL not null primary key,
    user_id int not null references users,
    content varchar not null,
    spoiler_of varchar,
    nsfw boolean not null default false,
    likes_count int not null default 0,
    comments_count int not null default 0,
    create_at timestamp not null default now()
    
);
create index sorted_posts on posts (create_at DESC);

create table timeline (
	id serial not null primary key,
    user_id int not null references users,
    post_id int not null references posts
);
create unique index timeline_unique on timeline (user_id,post_id);

create table post_likes (
    user_id int not null references users,
    post_id int not null references posts,
    primary key(user_id,post_id)
);

create table comments (
	id SERIAL not null primary key,
    user_id int not null references users,
    post_id int not null references posts,
    content varchar not null,
    likes_count int not null default 0,
    create_at timestamp not null default now()
    
);

create table  post_subcriptions (
    user_id int not null references users,
    post_id int not null references posts,
    primary key(user_id,post_id)
);

create index sorted_comments on comments (create_at DESC);


create table comment_likes (
    user_id int not null references users,
    comment_id int not null references comments,
    primary key(user_id,comment_id)
);

create table notifications (
    id serial not null primary key,
    user_id int not null references users,
    post_id int  references posts,
    actors VARCHAR[] not null,
    type varchar not null,
    read boolean not null default false,
    issued_at timestamp not null default now()
);
create index sorted_notifications on notifications (issued_at DESC);
create   UNIQUE INDEX unique_notifications on notifications (user_id, type, post_id, read);

INSERT INTO users  (id,email,username) VALUES (1,'john@dot.com','john');
INSERT INTO users  (id,email,username) VALUES (2,'jane@dot.com','jane');
INSERT INTO posts (id,user_id,content,comments_count) values (1,1,'test',1);
INSERT INTO timeline (id,user_id,post_id) values (1,1,1);
INSERT INTO comments (id,user_id,post_id,content) values (1,1,1,'comment test');
INSERT INTO post_subcriptions (user_id,post_id) values (1,1);