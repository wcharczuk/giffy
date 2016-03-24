begin;

select create_column('vote_summary', 'created_utc', '

alter table vote_summary add created_utc timestamp;

update 
    vote_summary
set
    created_utc = i.created_utc
from 
    image i
where
    vote_summary.image_id = i.id
    and vote_summary.created_utc is null;
    
alter table vote_summary alter created_utc set not null;
');

commit;
