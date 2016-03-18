begin;

select create_table('moderation', 'CREATE TABLE moderation (
	user_id bigint not null,
	uuid varchar(32) not null,
	timestamp_utc timestamp not null,
	verb varchar(32) not null,
	noun varchar(255) not null,
	secondary_noun varchar(255)
);
ALTER TABLE moderation ADD CONSTRAINT pk_moderation_uuid PRIMARY KEY (uuid);
ALTER TABLE moderation add CONSTRAINT fk_moderation_user_id FOREIGN KEY (user_id) REFERENCES users(id);');