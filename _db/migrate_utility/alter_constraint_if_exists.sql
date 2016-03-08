DROP FUNCTION IF EXISTS alter_constraint_if_exists(text,text);
CREATE FUNCTION alter_constraint_if_exists(constraint_name text, statement text) RETURNS void AS
$_$
BEGIN

IF EXISTS (
		SELECT 	1
		FROM	pg_constraint
		WHERE	conname = constraint_name
	) THEN
	EXECUTE statement;
	RAISE NOTICE 'MIGRATE :: ALTERED CONSTRAINT %', constraint_name;
END IF;

END;
$_$ LANGUAGE plpgsql;