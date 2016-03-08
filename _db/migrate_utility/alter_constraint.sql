DROP FUNCTION IF EXISTS alter_constraint(text,text);
CREATE FUNCTION alter_constraint(constraint_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
		SELECT 	1
		FROM	pg_constraint
		WHERE	conname = constraint_name
	) THEN
	EXECUTE statement;
	RAISE NOTICE 'MIGRATE :: ALTERED CONSTRAINT %', constraint_name;
ELSE
  RAISE EXCEPTION 'MIGRATE :: CONSTRAINT % DOES NOT EXIST', constraint_name;
END IF;

END;
$_$ LANGUAGE plpgsql;