CREATE TABLE user (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	username varchar(255) not null,
	first_name varchar(64),
	last_name varchar(64)
);
ALTER TABLE user ADD CONSTRAINT pk_user_id PRIMARY KEY (id);
ALTER TABLE user ADD CONSTRAINT uk_user_uuid UNIQUE (uuid);
ALTER TABLE user ADD CONSTRAINT uk_user_username UNIQUE (username);

CREATE TABLE image (
	id serial not null,
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,
	updated_utc timestamp,
	updated_by bigint,

	display_name varchar(64),

	md5 bytea not null,
	s3_read_url string varchar(1024),
	s3_bucket varchar(32) not null,
	s3_key varchar(32) not null,

	width int not null,
	height int not null,
	extension varchar(5)
);
ALTER TABLE image ADD CONSTRAINT pk_image_id PRIMARY KEY (id);
ALTER TABLE image ADD CONSTRAINT uk_image_uuid UNIQUE (uuid);
ALTER TABLE image ADD CONSTRAINT fk_image_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES user(id);
ALTER TABLE image ADD CONSTRAINT fk_image_updated_by_user_id 
	FOREIGN KEY (updated_by) REFERENCES user(id);

CREATE TABLE image_tag (
	uuid varchar(32) not null,
	created_utc timestamp not null,
	created_by bigint not null,
	image_id bigint not null,
	tag_value varchar(32) not null
);
ALTER TABLE image_tag ADD CONSTRAINT pk_image_tag_uuid PRIMARY KEY (id);
ALTER TABLE image_tag ADD CONSTRAINT fk_image_tag_created_by_user_id 
	FOREIGN KEY (created_by) REFERENCES user(id);
