create extension if not exists pgcrypto;
create extension if not exists pg_trgm;

CREATE TABLE users (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	username varchar(255) not null,
	first_name varchar(64),
	last_name varchar(64),
    email_address varchar(255),
    is_email_verified boolean not null,
    is_admin boolean not null default false,
    is_moderator boolean not null default false,
	is_banned boolean not null default false
);
ALTER TABLE users ADD CONSTRAINT pk_users_id PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT uk_users_uuid UNIQUE (uuid);
ALTER TABLE users ADD CONSTRAINT uk_users_username UNIQUE (username);

CREATE TABLE user_auth (
    user_id bigint not null,
    provider varchar(32) not null,
    timestamp_utc timestamp not null,
    auth_token bytea not null,
	auth_token_hash bytea not null,
    auth_secret bytea
);
ALTER TABLE user_auth ADD CONSTRAINT pk_user_auth_user_id_provider PRIMARY KEY (user_id,provider);
ALTER TABLE user_auth ADD CONSTRAINT fk_user_auth_user_id FOREIGN KEY (user_id) REFERENCES users(id);
CREATE INDEX ix_user_auth_auth_token_hash ON user_auth(auth_token_hash);

CREATE TABLE user_session (
	session_id varchar(32) not null,
    user_id bigint not null,
    timestamp_utc timestamp not null
);
ALTER TABLE user_session ADD CONSTRAINT pk_user_session_session_id PRIMARY KEY (session_id);
ALTER TABLE user_session ADD CONSTRAINT fk_user_session_user_id FOREIGN KEY (user_id) REFERENCES users(id);

CREATE TABLE image (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,

	display_name varchar(255),

	md5 bytea not null,
	s3_read_url varchar(1024),
	s3_bucket varchar(64) not null,
	s3_key varchar(64) not null,

	width int not null,
	height int not null,
	file_size int not null,
	
	extension varchar(8)
);
ALTER TABLE image ADD CONSTRAINT pk_image_id PRIMARY KEY (id);
ALTER TABLE image ADD CONSTRAINT uk_image_uuid UNIQUE (uuid);
ALTER TABLE image ADD CONSTRAINT uk_image_md5 UNIQUE (md5);
ALTER TABLE image ADD CONSTRAINT fk_image_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES users(id);

CREATE TABLE tag (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,
	tag_value varchar(32) not null
);
ALTER TABLE tag ADD CONSTRAINT pk_tag_id PRIMARY KEY (id);
ALTER TABLE tag ADD CONSTRAINT uk_tag_uuid UNIQUE (uuid);
ALTER TABLE tag ADD CONSTRAINT uk_tag_tag_value UNIQUE (tag_value);
ALTER TABLE tag ADD CONSTRAINT fk_tag_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES users(id);

CREATE TABLE vote_summary (
	image_id bigint not null,
	tag_id bigint not null,
	last_vote_by bigint not null,
	last_vote_utc timestamp not null,
	votes_for int not null,
	votes_against int not null,
	votes_total int not null
);
ALTER TABLE vote_summary ADD CONSTRAINT pk_vote_summary_image_id_tag_id PRIMARY KEY (image_id, tag_id);
ALTER TABLE vote_summary ADD CONSTRAINT fk_vote_summary_image_id
	FOREIGN KEY (image_id) REFERENCES image(id);
ALTER TABLE vote_summary ADD CONSTRAINT fk_vote_summary_tag_id 
	FOREIGN KEY (tag_id) REFERENCES tag(id);

CREATE TABLE vote (
	user_id bigint not null,
	image_id bigint not null,
	tag_id bigint not null,
	timestamp_utc timestamp not null,
	is_upvote bool not null
);
ALTER TABLE vote ADD CONSTRAINT pk_vote_user_id_image_id_tag_id PRIMARY KEY (user_id, image_id, tag_id);
ALTER TABLE vote ADD CONSTRAINT fk_vote_user_id
	FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE vote ADD CONSTRAINT fk_vote_image_id
	FOREIGN KEY (image_id) REFERENCES image(id);
ALTER TABLE vote ADD CONSTRAINT fk_vote_tag_id
	FOREIGN KEY (tag_id) REFERENCES tag(id);

CREATE TABLE moderation (
	user_id bigint not null,
	uuid varchar(32) not null,
	timestamp_utc timestamp not null,
	verb varchar(32) not null,
	object varchar(255) not null,
	noun varchar(255),
	secondary_noun varchar(255)
);
ALTER TABLE moderation ADD CONSTRAINT pk_moderation_uuid PRIMARY KEY (uuid);
ALTER TABLE moderation add CONSTRAINT fk_moderation_user_id FOREIGN KEY (user_id) REFERENCES users(id);