begin;

select create_table('search_history', '
create table search_history (
	source varchar(255) not null,
	source_user_identifier varchar(255),
	source_team_identifier varchar(255),
	source_channel_identifier varchar(255),

	timestamp_utc timestamp not null,
	search_query varchar(255) not null,

	did_find_match boolean not null,

	image_id bigint,
	tag_id bigint
);
');

commit;