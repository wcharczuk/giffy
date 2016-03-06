DROP FUNCTION IF EXISTS alter_index_if_exists(text,text,text);
CREATE FUNCTION alter_index_if_exists(table_name text, index_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
      SELECT  1
      FROM  pg_catalog.pg_index ix  
        join  pg_catalog.pg_class t on t.oid = ix.indrelid
        join  pg_catalog.pg_class i on i.oid = ix.indexrelid
      WHERE 
        t.relname = table_name 
        and i.relname = index_name
        and t.relkind = 'r'
    ) THEN
    EXECUTE statement;
    RAISE NOTICE 'MIGRATE :: INDEX % ON TABLE %', index_name, table_name;
END IF;

END;
$_$ LANGUAGE plpgsql;