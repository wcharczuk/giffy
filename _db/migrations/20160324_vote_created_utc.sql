begin;

select alter_column('vote', 'timestamp_utc', '
    ALTER TABLE vote RENAME timestamp_utc TO created_utc;
');

commit;