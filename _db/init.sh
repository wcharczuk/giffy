#!/bin/sh

set -e

if [ ! -z "$GIFFY_APP" ]; then 
    : ${GIFFY_HOST:?"GIFFY_HOST is requried"}
    
    ssh dokku@${GIFFY_HOST} postgres:connect ${GIFFY_APP} < ./_db/init/02_schema.sql;
    ssh dokku@${GIFFY_HOST} postgres:connect ${GIFFY_APP} < ./_db/init/03_ref_data.sql;
    
     # FOR EACH FILE IN ./migrate_utility (these are necessary for migrations)
    for file in $(find ./_db/migrate_utility -name "*.sql" -type f);
    do
        ssh dokku@${GIFFY_HOST} postgres:connect ${GIFFY_APP} < ${file};
    done
else
    : ${DB_SCHEMA:?"DB_SCHEMA is required"}
    : ${DB_USER:?"DB_USER is required"}
    : ${DB_PASSWORD:?"DB_PASSWORD is required"}
    
    # use postgres as the template database ...
    psql postgres -f ./_db/init/01_create_db.sql;
    
    # rest assumes we have stood up the destination database
    psql $DB_SCHEMA -f ./_db/init/02_schema.sql;
    psql $DB_SCHEMA -f ./_db/init/03_ref_data.sql;

    # FOR EACH FILE IN ./migrate_utility (these are necessary for migrations)
    for file in $(find ./_db/migrate_utility -name "*.sql" -type f);
    do
        psql $DB_SCHEMA -f ${file};
    done
fi


