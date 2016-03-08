DROP FUNCTION IF EXISTS create_table(text,text);
CREATE FUNCTION create_table(table_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
      SELECT 	1
      FROM	pg_catalog.pg_tables
      WHERE	tablename  = table_name
    ) THEN
   RAISE NOTICE 'MIGRATE :: TABLE % ALREADY EXISTS', table_name;
ELSE
   EXECUTE statement;
   RAISE NOTICE 'MIGRATE :: CREATED TABLE %', table_name;
END IF;

END;
$_$ LANGUAGE plpgsql;