DROP FUNCTION IF EXISTS create_constraint(text,text);
CREATE FUNCTION create_constraint(constraint_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
      SELECT 	1
      FROM	pg_constraint
      WHERE	conname = constraint_name
    ) THEN
   RAISE NOTICE 'MIGRATE :: CONSTRAINT % ALREADY EXISTS', constraint_name;
ELSE
   EXECUTE statement;
   RAISE NOTICE 'MIGRATE :: CREATED CONSTRAINT %', constraint_name;
END IF;

END;
$_$ LANGUAGE plpgsql;