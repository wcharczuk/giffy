begin;

select create_column('users', 'is_banned', '
	ALTER TABLE users ADD is_banned boolean not null default false;
');

commit;