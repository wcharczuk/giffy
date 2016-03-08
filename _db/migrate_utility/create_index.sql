DROP FUNCTION IF EXISTS create_index(text,text,text);
CREATE FUNCTION create_index(table_name text, index_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
      SELECT 	1
      FROM	pg_catalog.pg_index ix	
        join  pg_catalog.pg_class t on t.oid = ix.indrelid
        join  pg_catalog.pg_class i on i.oid = ix.indexrelid
      WHERE	
      	t.relname = table_name 
      	and i.relname = index_name
      	and t.relkind = 'r'
    ) THEN
   RAISE NOTICE 'MIGRATE :: INDEX % ON TABLE % ALREADY EXISTS', index_name, table_name;
ELSE
   EXECUTE statement;
   RAISE NOTICE 'MIGRATE :: CREATED INDEX % ON TABLE %', index_name, table_name;
END IF;

END;
$_$ LANGUAGE plpgsql;