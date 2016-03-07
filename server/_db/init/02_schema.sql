CREATE TABLE users (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	username varchar(255) not null,
	first_name varchar(64),
	last_name varchar(64)
);
ALTER TABLE users ADD CONSTRAINT pk_users_id PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT uk_users_uuid UNIQUE (uuid);
ALTER TABLE users ADD CONSTRAINT uk_users_username UNIQUE (username);

CREATE TABLE image (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,
	updated_utc timestamp,
	updated_by bigint,

	display_name varchar(64),

	md5 bytea not null,
	s3_read_url varchar(1024),
	s3_bucket varchar(32) not null,
	s3_key varchar(32) not null,

	width int not null,
	height int not null,
	extension varchar(5)
);
ALTER TABLE image ADD CONSTRAINT pk_image_id PRIMARY KEY (id);
ALTER TABLE image ADD CONSTRAINT uk_image_uuid UNIQUE (uuid);
ALTER TABLE image ADD CONSTRAINT fk_image_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES users(id);
ALTER TABLE image ADD CONSTRAINT fk_image_updated_by_user_id 
	FOREIGN KEY (updated_by) REFERENCES users(id);

CREATE TABLE tag (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,
	tag_value varchar(32) not null
);
ALTER TABLE tag ADD CONSTRAINT pk_tag_id PRIMARY KEY (id);
ALTER TABLE tag ADD CONSTRAINT uk_tag_uuid_uuid UNIQUE (uuid);
ALTER TABLE tag ADD CONSTRAINT fk_tag_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES users(id);

CREATE TABLE image_tag_votes (
	image_id bigint not null,
	tag_id bigint not null,
	last_vote_by bigint not null,
	last_vote_utc timestamp not null,
	votes_for int not null,
	votes_against int not null,
	votes_total int not null
);
ALTER TABLE image_tag_votes ADD CONSTRAINT pk_image_tag_votes_image_id_tag_id PRIMARY KEY (image_id, tag_id);
ALTER TABLE image_tag_votes ADD CONSTRAINT fk_image_tag_votes_image_id
	FOREIGN KEY (image_id) REFERENCES image(id);
ALTER TABLE image_tag_votes ADD CONSTRAINT fk_image_tag_votes_tag_id 
	FOREIGN KEY (tag_id) REFERENCES tag(id);

CREATE TABLE vote_log (
	user_id bigint not null,
	image_id bigint not null,
	tag_id bigint not null,
	timestamp_utc timestamp not null,
	is_upvote bool not null
);
ALTER TABLE vote_log ADD CONSTRAINT pk_vote_log_image_id_tag_id PRIMARY KEY (image_id, tag_id);
ALTER TABLE vote_log ADD CONSTRAINT fk_vote_log_user_id
	FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE vote_log ADD CONSTRAINT fk_vote_log_image_id
	FOREIGN KEY (image_id) REFERENCES image(id);
ALTER TABLE vote_log ADD CONSTRAINT fk_vote_log_tag_id
	FOREIGN KEY (tag_id) REFERENCES tag(id);