begin:

select create_column('image', 'file_size', '
	alter table image add file_size int not null default 0;
');

commit;