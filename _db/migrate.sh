#! /bin/sh

if [ ! -z "$GIFFY_APP" ]; then 
    : ${GIFFY_HOST:?"GIFFY_HOST is requried"}
    
    for file in $(find ./_db/migrations -name "*.sql" -type f);
    do
        ssh dokku@${GIFFY_HOST} postgres:connect ${GIFFY_APP} < ${file}  
    done
else
    : ${DB_SCHEMA:?"DB_SCHEMA is required"}
    
    for file in $(find ./_db/migrations -name "*.sql" -type f);
    do
        psql $DB_SCHEMA -f ${file};
    done
fi