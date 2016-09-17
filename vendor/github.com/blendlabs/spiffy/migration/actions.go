package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

const (
	verbCreate = "create"
	verbAlter  = "alter"
	verbRun    = "run"

	nounColumn     = "column"
	nounTable      = "table"
	nounIndex      = "index"
	nounConstraint = "constraint"
	nounAlways     = "always"
)

// actionName joins a noun and a verb
func actionName(verb, noun string) string {
	return fmt.Sprintf("%v %v", verb, noun)
}

// guard1 is for guards that require (1) arg such as `create table` and create constraint`
type guard1 func(c *spiffy.DbConnection, tx *sql.Tx, arg string) (bool, error)

// guard2 is for guards that require (2) args such as `create column` and `create index`
type guard2 func(c *spiffy.DbConnection, tx *sql.Tx, arg1, arg2 string) (bool, error)

func actionImpl(verb, noun string, stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable) error {
	action := actionName(verb, noun)
	newStack := append(stack, action)
	err := body.Invoke(c, tx)

	if err != nil {
		if l != nil {
			return l.Errorf(newStack, err)
		}
		return nil
	}
	if l != nil {
		return l.Applyf(newStack, "done")
	}
	return nil
}

func action1impl(verb, noun string, guard guard1, guardArgName string, stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	action := actionName(verb, noun)
	newStack := append(stack, action)
	if len(args) < 1 {
		if l != nil {
			return l.Errorf(newStack, exception.Newf("`%s` requires (1) argument => %s", guardArgName))
		}
		return nil
	}
	subject := args[0]
	if exists, err := guard(c, tx, subject); err != nil {
		if l != nil {
			return l.Errorf(newStack, err)
		}
		return nil
	} else if (verb == verbCreate && !exists) || (verb == verbAlter && exists) {
		err = body.Invoke(c, tx)
		if err != nil {
			if l != nil {
				return l.Errorf(newStack, err)
			}
			return nil
		}
		if l != nil {
			return l.Applyf(newStack, "`%s`", subject)
		}
		return nil
	}
	if l != nil {
		return l.Skipf(newStack, "`%s`", subject)
	}
	return nil
}

func action2impl(verb, noun string, guard guard2, guardArgNames, stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	action := actionName(verb, noun)
	newStack := append(stack, action)
	if len(args) < 2 {
		err := exception.Newf("`%s` requires (2) arguments => %s", strings.Join(guardArgNames, ", "))
		if l != nil {
			return l.Errorf(newStack, err)
		}
		return err
	}
	subject1 := args[0]
	subject2 := args[1]

	if exists, err := guard(c, tx, subject1, subject2); err != nil {
		if l != nil {
			return l.Errorf(newStack, err)
		}
		return err
	} else if (verb == verbCreate && !exists) || (verb == verbAlter && exists) {
		err = body.Invoke(c, tx)
		if err != nil {
			if l != nil {
				return l.Errorf(newStack, err)
			}
			return err
		}
		if l != nil {
			return l.Applyf(newStack, "`%s` on `%s`", subject2, subject1)
		}
		return nil
	}
	if l != nil {
		return l.Skipf(newStack, "`%s` on `%s`", subject2, subject1)
	}
	return nil
}

// --------------------------------------------------------------------------------
// Actions
// --------------------------------------------------------------------------------

// AlwaysRun always runs a step.
func AlwaysRun(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return actionImpl(verbRun, nounAlways, stack, l, c, tx, body)
}

// CreateColumn creates a table on the given connection if it does not exist.
func CreateColumn(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action2impl(verbCreate, nounColumn, ColumnExists, []string{"table_name", "column_name"}, stack, l, c, tx, body, args...)
}

// CreateConstraint creates a table on the given connection if it does not exist.
func CreateConstraint(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action1impl(verbCreate, nounConstraint, ConstraintExists, "constraint_name", stack, l, c, tx, body, args...)
}

// CreateTable creates a table on the given connection if it does not exist.
func CreateTable(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action1impl(verbCreate, nounTable, TableExists, "table_name", stack, l, c, tx, body, args...)
}

// CreateIndex creates a index on the given connection if it does not exist.
func CreateIndex(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action2impl(verbCreate, nounIndex, IndexExists, []string{"table_name", "index_name"}, stack, l, c, tx, body, args...)
}

// AlterColumn alters an existing column, erroring if it doesn't exist
func AlterColumn(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action2impl(verbAlter, nounTable, ColumnExists, []string{"table_name", "column_name"}, stack, l, c, tx, body, args...)
}

// AlterConstraint alters an existing constraint, erroring if it doesn't exist
func AlterConstraint(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action1impl(verbAlter, nounConstraint, ConstraintExists, "constraint_name", stack, l, c, tx, body, args...)
}

// AlterTable alters an existing table, erroring if it doesn't exist
func AlterTable(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action1impl(verbAlter, nounTable, TableExists, "table_name", stack, l, c, tx, body, args...)
}

// AlterIndex alters an existing index, erroring if it doesn't exist
func AlterIndex(stack []string, l *Logger, c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return action2impl(verbAlter, nounIndex, IndexExists, []string{"table_name", "index_name"}, stack, l, c, tx, body, args...)
}

// --------------------------------------------------------------------------------
// Execute
// --------------------------------------------------------------------------------

// Execute runs a given statement. You should guard the statement for re-runability yourself.
func Execute(c *spiffy.DbConnection, tx *sql.Tx, body Invocable, args ...string) error {
	return body.Invoke(c, tx)
}

// --------------------------------------------------------------------------------
// Guards
// --------------------------------------------------------------------------------

// TableExists returns if a table exists on the given connection.
func TableExists(c *spiffy.DbConnection, tx *sql.Tx, tableName string) (bool, error) {
	return c.QueryInTransaction(`SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1`, tx, strings.ToLower(tableName)).Any()
}

// ColumnExists returns if a column exists on a table on the given connection.
func ColumnExists(c *spiffy.DbConnection, tx *sql.Tx, tableName, columnName string) (bool, error) {
	return c.QueryInTransaction(`SELECT 1 FROM information_schema.columns i WHERE i.table_name = $1 and i.column_name = $2`, tx, strings.ToLower(tableName), strings.ToLower(columnName)).Any()
}

// ConstraintExists returns if a constraint exists on a table on the given connection.
func ConstraintExists(c *spiffy.DbConnection, tx *sql.Tx, constraintName string) (bool, error) {
	return c.QueryInTransaction(`SELECT 1 FROM pg_constraint WHERE conname = $1`, tx, strings.ToLower(constraintName)).Any()
}

// IndexExists returns if a index exists on a table on the given connection.
func IndexExists(c *spiffy.DbConnection, tx *sql.Tx, tableName, indexName string) (bool, error) {
	return c.QueryInTransaction(`SELECT 1 FROM pg_catalog.pg_index ix join pg_catalog.pg_class t on t.oid = ix.indrelid join pg_catalog.pg_class i on i.oid = ix.indexrelid WHERE t.relname = $1 and i.relname = $2 and t.relkind = 'r'`, tx, strings.ToLower(tableName), strings.ToLower(indexName)).Any()
}
