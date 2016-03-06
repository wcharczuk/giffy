DROP FUNCTION IF EXISTS create_column(text,text,text);
CREATE FUNCTION create_column(table_name_param text, column_name_param text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
    SELECT 	1
    FROM	information_schema.columns i
    WHERE	i.table_name  = table_name_param and i.column_name = column_name_param
  ) THEN
	RAISE NOTICE 'MIGRATE :: COLUMN % ON TABLE % ALREADY EXISTS', column_name_param, table_name_param;
ELSE
	EXECUTE statement;
	RAISE NOTICE 'MIGRATE :: CREATED COLUMN % ON TABLE %', column_name_param, table_name_param;
END IF;

END;
$_$ LANGUAGE plpgsql;