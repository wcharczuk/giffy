package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/stringutil"
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

// GuardFunc is a control for migration steps.
type GuardFunc func(*Suite, *Group, *Step, *db.Connection, *sql.Tx) error

// --------------------------------------------------------------------------------
// Guards
// --------------------------------------------------------------------------------

// Guard returns a function that determines if a step in a group should run.
func Guard(description string, evaluator func(c *db.Connection, tx *sql.Tx) (bool, error)) GuardFunc {
	return func(suite *Suite, group *Group, step *Step, c *db.Connection, tx *sql.Tx) error {
		proceed, err := evaluator(c, tx)
		if err != nil {
			if suite != nil {
				return suite.error(group, step, err)
			}
			return err
		}

		if !proceed {
			if suite != nil {
				suite.skipf(group, step, description)
			}
			return nil
		}

		err = step.body(c, tx)
		if err != nil {
			if suite != nil {
				return suite.error(group, step, err)
			}
			return err
		}
		if suite != nil {
			suite.applyf(group, step, description)
		}
		return nil
	}
}

// AlwaysRun always runs a step.
func AlwaysRun() GuardFunc {
	return Guard("always run", func(_ *db.Connection, _ *sql.Tx) (bool, error) { return true, nil })
}

// IfExists only runs the statement if the given item exists.
func IfExists(statement string) GuardFunc {
	return Guard("if exists run", func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return exists(c, tx, statement)
	})
}

// IfNotExists only runs the statement if the given item doesn't exist.
func IfNotExists(statement string) GuardFunc {
	return Guard("if not exists run", func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return notExists(c, tx, statement)
	})
}

// ColumnNotExists creates a table on the given connection if it does not exist.
func ColumnNotExists(tableName, columnName string) GuardFunc {
	return Guard(fmt.Sprintf("create column `%s` on `%s`", columnName, tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return not(columnExists(c, tx, tableName, columnName))
	})
}

// ConstraintNotExists creates a table on the given connection if it does not exist.
func ConstraintNotExists(constraintName string) GuardFunc {
	return Guard(fmt.Sprintf("create constraint `%s`", constraintName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return not(constraintExists(c, tx, constraintName))
	})
}

// TableNotExists creates a table on the given connection if it does not exist.
func TableNotExists(tableName string) GuardFunc {
	return Guard(fmt.Sprintf("create table `%s`", tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return not(tableExists(c, tx, tableName))
	})
}

// IndexNotExists creates a index on the given connection if it does not exist.
func IndexNotExists(tableName, indexName string) GuardFunc {
	return Guard(fmt.Sprintf("create index `%s` on `%s`", indexName, tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return not(indexExists(c, tx, tableName, indexName))
	})
}

// RoleNotExists creates a new role if it doesn't exist.
func RoleNotExists(roleName string) GuardFunc {
	return Guard(fmt.Sprintf("create role `%s`", roleName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return not(roleExists(c, tx, roleName))
	})
}

// ColumnExists alters an existing column, erroring if it doesn't exist
func ColumnExists(tableName, columnName string) GuardFunc {
	return Guard(fmt.Sprintf("alter column `%s` on `%s`", columnName, tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return columnExists(c, tx, tableName, columnName)
	})
}

// ConstraintExists alters an existing constraint, erroring if it doesn't exist
func ConstraintExists(constraintName string) GuardFunc {
	return Guard(fmt.Sprintf("alter constraint `%s`", constraintName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return constraintExists(c, tx, constraintName)
	})
}

// TableExists alters an existing table, erroring if it doesn't exist
func TableExists(tableName string) GuardFunc {
	return Guard(fmt.Sprintf("alter table `%s`", tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return tableExists(c, tx, tableName)
	})
}

// IndexExists alters an existing index, erroring if it doesn't exist
func IndexExists(tableName, indexName string) GuardFunc {
	return Guard(fmt.Sprintf("alter index `%s` on `%s`", indexName, tableName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return indexExists(c, tx, tableName, indexName)
	})
}

// RoleExists alters an existing role in the db
func RoleExists(roleName string) GuardFunc {
	return Guard(fmt.Sprintf("alter role `%s`", roleName), func(c *db.Connection, tx *sql.Tx) (bool, error) {
		return roleExists(c, tx, roleName)
	})
}

func not(proceed bool, err error) (bool, error) {
	return !proceed, err
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
	if !stringutil.HasPrefixCaseless(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).Any()
}

// notExists returns if a statement doesnt have results.
func notExists(c *db.Connection, tx *sql.Tx, selectStatement string) (bool, error) {
	if !stringutil.HasPrefixCaseless(selectStatement, "select") {
		return false, fmt.Errorf("statement must be a `SELECT`")
	}
	return c.QueryInTx(selectStatement, tx).None()
}
