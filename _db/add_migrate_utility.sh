#! /bin/sh

: ${GIFFY_APP:?"GIFFY_APP is required"}
: ${GIFFY_HOST:?"GIFFY_HOST is requried"}

# FOR EACH FILE IN ./migrate_utility (these are necessary for migrations)
for file in $(find ./_db/migrate_utility -name "*.sql" -type f);
do
    ssh dokku@${GIFFY_HOST} postgres:connect ${GIFFY_APP} < ${file}
done