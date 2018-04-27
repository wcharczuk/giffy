package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/util"
)

const (
	verbCreate = "create"
	verbAlter  = "alter"
	verbRun    = "run"

	nounColumn     = "column"
	nounTable      = "table"
	nounIndex      = "index"
	nounConstraint = "constraint"
	nounRole       = "role"

	adverbAlways    = "always"
	adverbExists    = "exists"
	adverbNotExists = "not exists"
)

// --------------------------------------------------------------------------------
// Guards
// --------------------------------------------------------------------------------

// DynamicGuard is guard that can dynamically determine if the inner invocable should run.
func DynamicGuard(label string, guard func(c *db.Connection, tx *sql.Tx) (bool, error)) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		s.WithLabel(label)

		proceed, err := guard(c, tx)
		if err != nil {
			return s.Collector().Error(s, err)
		}

		if proceed {
			err = s.body.Invoke(c, tx)
			if err != nil {
				return s.Collector().Error(s, err)
			}
			s.Collector().Applyf(s, label)
			return nil
		}

		s.Collector().Skipf(s, label)
		return nil
	}
}

// AlwaysRun always runs a step.
func AlwaysRun() Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl(s, verbRun, adverbAlways, c, tx)
	}
}

// IfExists only runs the statement if the given item exists.
func IfExists(statement string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbRun, adverbExists, exists, statement, c, tx)
	}
}

// IfNotExists only runs the statement if the given item doesn't exist.
func IfNotExists(statement string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbRun, adverbNotExists, notExists, statement, c, tx)
	}
}

// ColumnNotExists creates a table on the given connection if it does not exist.
func ColumnNotExists(tableName, columnName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl2(s, verbCreate, nounColumn, columnExists, tableName, columnName, c, tx)
	}
}

// ConstraintNotExists creates a table on the given connection if it does not exist.
func ConstraintNotExists(constraintName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbCreate, nounConstraint, constraintExists, constraintName, c, tx)
	}
}

// TableNotExists creates a table on the given connection if it does not exist.
func TableNotExists(tableName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbCreate, nounTable, tableExists, tableName, c, tx)
	}
}

// IndexNotExists creates a index on the given connection if it does not exist.
func IndexNotExists(tableName, indexName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl2(s, verbCreate, nounIndex, indexExists, tableName, indexName, c, tx)
	}
}

// RoleNotExists creates a new role if it doesn't exist.
func RoleNotExists(roleName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbCreate, nounRole, roleExists, roleName, c, tx)
	}
}

// ColumnExists alters an existing column, erroring if it doesn't exist
func ColumnExists(tableName, columnName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl2(s, verbAlter, nounTable, columnExists, tableName, columnName, c, tx)
	}
}

// ConstraintExists alters an existing constraint, erroring if it doesn't exist
func ConstraintExists(constraintName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbAlter, nounConstraint, constraintExists, constraintName, c, tx)
	}
}

// TableExists alters an existing table, erroring if it doesn't exist
func TableExists(tableName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbAlter, nounTable, tableExists, tableName, c, tx)
	}
}

// IndexExists alters an existing index, erroring if it doesn't exist
func IndexExists(tableName, indexName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl2(s, verbAlter, nounIndex, indexExists, tableName, indexName, c, tx)
	}
}

// RoleExists alters an existing role in the db
func RoleExists(roleName string) Guard {
	return func(s *Step, c *db.Connection, tx *sql.Tx) error {
		return guardImpl1(s, verbAlter, nounRole, roleExists, roleName, c, tx)
	}
}

// actionName joins a noun and a verb
func actionName(verb, noun string) string {
	return fmt.Sprintf("%v %v", verb, noun)
}

// guard1 is for guards that require (1) arg such as `create table` and create constraint`
type guard1 func(c *db.Connection, tx *sql.Tx, arg string) (bool, error)

// guard2 is for guards that require (2) args such as `create column` and `create index`
type guard2 func(c *db.Connection, tx *sql.Tx, arg1, arg2 string) (bool, error)

// actionImpl is an unguarded action, it doesn't care if something exists or doesn't
// it is a requirement of the operation to guard itself.
func guardImpl(s *Step, verb, noun string, c *db.Connection, tx *sql.Tx) error {
	err := s.body.Invoke(c, tx)

	if err != nil {
		return s.collector.Error(s, err)
	}
	if s.collector != nil {
		s.collector.Applyf(s, "done")
	}
	return nil
}

func guardImpl1(s *Step, verb, noun string, guard guard1, subject string, c *db.Connection, tx *sql.Tx) error {
	s.WithLabel(actionName(verb, noun))

	if exists, err := guard(c, tx, subject); err != nil {
		return s.collector.Error(s, err)
	} else if (verb == verbCreate && !exists) ||
		(verb == verbAlter && exists) ||
		(verb == verbRun && exists) {
		err = s.body.Invoke(c, tx)
		if err != nil {
			return s.collector.Error(s, err)
		}
		s.collector.Applyf(s, "%s `%s`", verb, subject)
		return nil
	}
	s.collector.Skipf(s, "%s `%s`", verb, subject)
	return nil
}

func guardImpl2(s *Step, verb, noun string, guard guard2, subject1, subject2 string, c *db.Connection, tx *sql.Tx) error {
	s.WithLabel(actionName(verb, noun))

	if exists, err := guard(c, tx, subject1, subject2); err != nil {
		return s.collector.Error(s, err)
	} else if (verb == verbCreate && !exists) || (verb == verbAlter && exists) || (verb == verbRun && exists) {
		err = s.body.Invoke(c, tx)
		if err != nil {
			return s.collector.Error(s, err)
		}

		s.collector.Applyf(s, "%s `%s` on `%s`", verb, subject2, subject1)
		return nil
	}

	s.collector.Skipf(s, "%s `%s` on `%s`", verb, subject2, subject1)
	return nil
}

// --------------------------------------------------------------------------------
// Guards Implementations
// --------------------------------------------------------------------------------

// TableExists returns if a table exists on the given connection.
func tableExists(c *db.Connection, tx *sql.Tx, tableName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = $1`, tx, strings.ToLower(tableName)).Any()
}

// ColumnExists returns if a column exists on a table on the given connection.
func columnExists(c *db.Connection, tx *sql.Tx, tableName, columnName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM information_schema.columns i WHERE i.table_name = $1 and i.column_name = $2`, tx, strings.ToLower(tableName), strings.ToLower(columnName)).Any()
}

// ConstraintExists returns if a constraint exists on a table on the given connection.
func constraintExists(c *db.Connection, tx *sql.Tx, constraintName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_constraint WHERE conname = $1`, tx, strings.ToLower(constraintName)).Any()
}

// IndexExists returns if a index exists on a table on the given connection.
func indexExists(c *db.Connection, tx *sql.Tx, tableName, indexName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_catalog.pg_index ix join pg_catalog.pg_class t on t.oid = ix.indrelid join pg_catalog.pg_class i on i.oid = ix.indexrelid WHERE t.relname = $1 and i.relname = $2 and t.relkind = 'r'`, tx, strings.ToLower(tableName), strings.ToLower(indexName)).Any()
}

// roleExists returns if a role exists or not.
func roleExists(c *db.Connection, tx *sql.Tx, roleName string) (bool, error) {
	return c.QueryInTx(`SELECT 1 FROM pg_roles WHERE rolname ilike $1`, tx, roleName).Any()
}

// exists returns if a statement has results.
func exists(c *db.Connection, tx *sql.Tx, selectStatement string) (bool, error) {
	if !util.String.HasPrefixCaseInsensitive(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).Any()
}

// notExists returns if a statement doesnt have results.
func notExists(c *db.Connection, tx *sql.Tx, selectStatement string) (bool, error) {
	if !util.String.HasPrefixCaseInsensitive(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).None()
}
