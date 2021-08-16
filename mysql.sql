use backend;

create table  users (
	id serial not null primary key,
    email varchar(50) not null unique,
    username varchar(50) not null unique,
    avartar varchar(50) not null,
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
    content varchar(255) not null,
    spoiler_of varchar(50),
    nsfw boolean not null,
    create_at timestamp not null default now()
    
);

create index sorted_posts on posts (create_at DESC);

create table timeline (
	id serial not null primary key,
    user_id int not null references users,
    post_id int not null references posts
);

create unique index timeline_unique on timeline (user_id,post_id);