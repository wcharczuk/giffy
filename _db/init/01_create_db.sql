\set database_name `echo $DB_NAME`
\set db_user `echo $DB_USER`
\set db_password `echo $DB_PASSWORD`

--mercilessly slaughter connections
SELECT
    pg_terminate_backend(pg_stat_activity.pid)
FROM
    pg_stat_activity
WHERE
    pg_stat_activity.datname = :'database_name'
    AND pid <> pg_backend_pid();

DROP DATABASE IF EXISTS :database_name;

CREATE USER :db_user WITH PASSWORD :'db_password';

ALTER USER :db_user WITH SUPERUSER;

CREATE DATABASE :database_name WITH OWNER :db_user;

GRANT ALL PRIVILEGES ON ALL TABLES in SCHEMA public TO :db_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES in SCHEMA public TO :db_user;