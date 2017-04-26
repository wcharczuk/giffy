package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/blendlabs/spiffy"
)

const (
	verbCreate = "create"
	verbAlter  = "alter"
	verbRun    = "run"

	nounColumn      = "column"
	nounTable       = "table"
	nounIndex       = "index"
	nounConstraint  = "constraint"
	nounRole        = "role"
	nounAlways      = "always"
	nounIfExists    = "if exists"
	nounIfNotExists = "if not exists"
)

// actionName joins a noun and a verb
func actionName(verb, noun string) string {
	return fmt.Sprintf("%v %v", verb, noun)
}

// guard1 is for guards that require (1) arg such as `create table` and create constraint`
type guard1 func(c *spiffy.Connection, tx *sql.Tx, arg string) (bool, error)

// guard2 is for guards that require (2) args such as `create column` and `create index`
type guard2 func(c *spiffy.Connection, tx *sql.Tx, arg1, arg2 string) (bool, error)

// actionImpl is an unguarded action, it doesn't care if something exists or doesn't
// it is a requirement of the operation to guard itself.
func actionImpl(o *Operation, verb, noun string, c *spiffy.Connection, tx *sql.Tx) error {
	err := o.body.Invoke(c, tx)

	if err != nil {
		if o.logger != nil {
			return o.logger.Errorf(o, err)
		}
		return nil
	}
	if o.logger != nil {
		return o.logger.Applyf(o, "done")
	}
	return nil
}

func actionImpl1(o *Operation, verb, noun string, guard guard1, guardArgName string, c *spiffy.Connection, tx *sql.Tx) error {
	o.SetLabel(actionName(verb, noun))
	if len(o.args) < 1 {
		err := fmt.Errorf("`%s` requires (1) argument => %s", o.label, guardArgName)
		if o.logger != nil {
			return o.logger.Errorf(o, err)
		}
		return err
	}
	subject := o.args[0]
	if exists, err := guard(c, tx, subject); err != nil {
		if o.logger != nil {
			return o.logger.Errorf(o, err)
		}
		return nil
	} else if (verb == verbCreate && !exists) ||
		(verb == verbAlter && exists) ||
		(verb == verbRun && exists) {
		err = o.body.Invoke(c, tx)
		if err != nil {
			if o.logger != nil {
				return o.logger.Errorf(o, err)
			}
			return nil
		}
		if o.logger != nil {
			return o.logger.Applyf(o, "%s `%s`", verb, subject)
		}
		return nil
	}
	if o.logger != nil {
		return o.logger.Skipf(o, "%s `%s`", verb, subject)
	}
	return nil
}

func actionImpl2(o *Operation, verb, noun string, guard guard2, guardArgNames []string, c *spiffy.Connection, tx *sql.Tx) error {
	o.SetLabel(actionName(verb, noun))
	if len(o.args) < 2 {
		err := fmt.Errorf("`%s` requires (2) arguments => %s", o.label, strings.Join(guardArgNames, ", "))
		if o.logger != nil {
			return o.logger.Errorf(o, err)
		}
		return err
	}
	subject1 := o.args[0]
	subject2 := o.args[1]

	if exists, err := guard(c, tx, subject1, subject2); err != nil {
		if o.logger != nil {
			return o.logger.Errorf(o, err)
		}
		return err
	} else if (verb == verbCreate && !exists) || (verb == verbAlter && exists) || (verb == verbRun && exists) {
		err = o.body.Invoke(c, tx)
		if err != nil {
			if o.logger != nil {
				return o.logger.Errorf(o, err)
			}
			return err
		}
		if o.logger != nil {
			return o.logger.Applyf(o, "%s `%s` on `%s`", verb, subject2, subject1)
		}
		return nil
	}
	if o.logger != nil {
		return o.logger.Skipf(o, "%s `%s` on `%s`", verb, subject2, subject1)
	}
	return nil
}

// --------------------------------------------------------------------------------
// Actions
// --------------------------------------------------------------------------------

// AlwaysRun always runs a step.
func AlwaysRun(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl(o, verbRun, nounAlways, c, tx)
}

// IfExists only runs the statement if the given item exists.
func IfExists(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbRun, nounIfExists, exists, "select_statement", c, tx)
}

// IfNotExists only runs the statement if the given item doesn't exist.
func IfNotExists(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbRun, nounIfNotExists, notExists, "select_statement", c, tx)
}

// CreateColumn creates a table on the given connection if it does not exist.
func CreateColumn(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl2(o, verbCreate, nounColumn, columnExists, []string{"table_name", "column_name"}, c, tx)
}

// CreateConstraint creates a table on the given connection if it does not exist.
func CreateConstraint(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbCreate, nounConstraint, constraintExists, "constraint_name", c, tx)
}

// CreateTable creates a table on the given connection if it does not exist.
func CreateTable(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbCreate, nounTable, tableExists, "table_name", c, tx)
}

// CreateIndex creates a index on the given connection if it does not exist.
func CreateIndex(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl2(o, verbCreate, nounIndex, indexExists, []string{"table_name", "index_name"}, c, tx)
}

// CreateRole creates a new role if it doesn't exist.
func CreateRole(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbCreate, nounRole, roleExists, "role_name", c, tx)
}

// AlterColumn alters an existing column, erroring if it doesn't exist
func AlterColumn(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl2(o, verbAlter, nounTable, columnExists, []string{"table_name", "column_name"}, c, tx)
}

// AlterConstraint alters an existing constraint, erroring if it doesn't exist
func AlterConstraint(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbAlter, nounConstraint, constraintExists, "constraint_name", c, tx)
}

// AlterTable alters an existing table, erroring if it doesn't exist
func AlterTable(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbAlter, nounTable, tableExists, "table_name", c, tx)
}

// AlterIndex alters an existing index, erroring if it doesn't exist
func AlterIndex(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl2(o, verbAlter, nounIndex, indexExists, []string{"table_name", "index_name"}, c, tx)
}

// AlterRole alters an existing role in the db
func AlterRole(o *Operation, c *spiffy.Connection, tx *sql.Tx) error {
	return actionImpl1(o, verbAlter, nounRole, roleExists, "role_name", c, tx)
}

// --------------------------------------------------------------------------------
// Guards
// --------------------------------------------------------------------------------

// TableExists returns if a table exists on the given connection.
func tableExists(c *spiffy.Connection, tx *sql.Tx, tableName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1`, tx, strings.ToLower(tableName)).Any()
}

// ColumnExists returns if a column exists on a table on the given connection.
func columnExists(c *spiffy.Connection, tx *sql.Tx, tableName, columnName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM information_schema.columns i WHERE i.table_name = $1 and i.column_name = $2`, tx, strings.ToLower(tableName), strings.ToLower(columnName)).Any()
}

// ConstraintExists returns if a constraint exists on a table on the given connection.
func constraintExists(c *spiffy.Connection, tx *sql.Tx, constraintName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_constraint WHERE conname = $1`, tx, strings.ToLower(constraintName)).Any()
}

// IndexExists returns if a index exists on a table on the given connection.
func indexExists(c *spiffy.Connection, tx *sql.Tx, tableName, indexName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_catalog.pg_index ix join pg_catalog.pg_class t on t.oid = ix.indrelid join pg_catalog.pg_class i on i.oid = ix.indexrelid WHERE t.relname = $1 and i.relname = $2 and t.relkind = 'r'`, tx, strings.ToLower(tableName), strings.ToLower(indexName)).Any()
}

// roleExists returns if a role exists or not.
func roleExists(c *spiffy.Connection, tx *sql.Tx, roleName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_roles WHERE rolname ilike $1`, tx, roleName).Any()
}

// exists returns if a statement has results.
func exists(c *spiffy.Connection, tx *sql.Tx, selectStatement string) (bool, error) {
	if !spiffy.HasPrefixCaseInsensitive(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).Any()
}

// notExists returns if a statement doesnt have results.
func notExists(c *spiffy.Connection, tx *sql.Tx, selectStatement string) (bool, error) {
	if !spiffy.HasPrefixCaseInsensitive(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).None()
}
