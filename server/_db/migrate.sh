#! /bin/sh

: ${DB_MIGRATE_MODE:?"DB_MIGRATE_MODE is required"}
: ${DB_SCHEMA:?"DB_SCHEMA is required"}

for file in $(find ./_db/migrations -not -path '*/_*' -name "*.sql" -type f);
do
    psql $DB_SCHEMA -f ${file};
    if [ "$DB_MIGRATE_MODE" = "release" ]; then
        mv $file ${file}.run_$(date +"%Y")_$(date +"%M")_$(date +"%d");
    fi
done