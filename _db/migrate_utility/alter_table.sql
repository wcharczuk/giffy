DROP FUNCTION IF EXISTS alter_table(text,text,text);
CREATE FUNCTION alter_table(input_table_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
	 	SELECT 	1
	    FROM	pg_catalog.pg_tables
	    WHERE	tablename  = input_table_name
	) THEN
	EXECUTE statement;
	RAISE NOTICE 'MIGRATE :: ALTERED TABLE %', input_table_name;
ELSE
	RAISE EXCEPTION 'MIGRATE :: TABLE % DOESNT EXIST', input_table_name;
END IF;

END;
$_$ LANGUAGE plpgsql;