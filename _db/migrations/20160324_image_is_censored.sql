begin;

select create_column('image','is_censored', '
    alter table image add is_censored boolean not null default false;
');

commit;