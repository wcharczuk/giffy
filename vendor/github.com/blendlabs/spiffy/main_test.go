package spiffy

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/blendlabs/go-exception"
)

//------------------------------------------------------------------------------------------------
// Testing Entrypoint
//------------------------------------------------------------------------------------------------

// TestMain is the testing entrypoint.
func TestMain(m *testing.M) {
	config := dbConnectionFromEnvironment()
	CreateDbAlias("main", config)
	SetDefaultAlias("main")

	os.Exit(m.Run())
}

func dbConnectionFromEnvironment() *DbConnection {
	dbHost := os.Getenv("DB_HOST")
	dbSchema := os.Getenv("DB_SCHEMA")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbHost == "" {
		dbHost = "localhost"
	}

	if dbSchema == "" {
		dbSchema = "postgres"
	}

	return NewDbConnection(dbHost, dbSchema, dbUser, dbPassword)
}

//------------------------------------------------------------------------------------------------
// Benchmarking
//------------------------------------------------------------------------------------------------

type benchObj struct {
	ID        int       `db:"id,pk,serial"`
	Name      string    `db:"name"`
	Timestamp time.Time `db:"timestamp_utc"`
	Amount    float32   `db:"amount"`
	Pending   bool      `db:"pending"`
	Category  string    `db:"category"`
}

func (b *benchObj) Populate(rows *sql.Rows) error {
	return rows.Scan(&b.ID, &b.Name, &b.Timestamp, &b.Amount, &b.Pending, &b.Category)
}

func (b benchObj) TableName() string {
	return "bench_object"
}

func createTable(tx *sql.Tx) error {
	createSQL := `CREATE TABLE IF NOT EXISTS bench_object (id serial not null, name varchar(255), timestamp_utc timestamp, amount real, pending boolean, category varchar(255));`
	return DefaultDb().ExecInTransaction(createSQL, tx)
}

func createObject(index int, tx *sql.Tx) error {
	obj := benchObj{}
	obj.Name = fmt.Sprintf("test_object_%d", index)
	obj.Timestamp = time.Now().UTC()
	obj.Amount = 1000.0 + (5.0 * float32(index))
	obj.Pending = index%2 == 0
	obj.Category = fmt.Sprintf("category_%d", index)
	return exception.Wrap(DefaultDb().CreateInTransaction(&obj, tx))
}

func seedObjects(count int, tx *sql.Tx) error {
	createTableErr := createTable(tx)
	if createTableErr != nil {
		return exception.Wrap(createTableErr)
	}

	for i := 0; i < count; i++ {
		createObjErr := createObject(i, tx)
		if createObjErr != nil {
			return exception.Wrap(createObjErr)
		}
	}
	return nil
}

func readManual(tx *sql.Tx) ([]benchObj, error) {
	objs := []benchObj{}
	readSQL := `select id,name,timestamp_utc,amount,pending,category from bench_object`
	readStmt, readStmtErr := DefaultDb().Prepare(readSQL, tx)
	if readStmtErr != nil {
		return nil, readStmtErr
	}
	defer readStmt.Close()

	rows, queryErr := readStmt.Query()
	defer rows.Close()
	if queryErr != nil {
		return nil, queryErr
	}

	for rows.Next() {
		obj := &benchObj{}
		populateErr := obj.Populate(rows)
		if populateErr != nil {
			return nil, populateErr
		}
		objs = append(objs, *obj)
	}

	return objs, nil
}

func readOrm(tx *sql.Tx) ([]benchObj, error) {
	objs := []benchObj{}
	allErr := DefaultDb().GetAllInTransaction(&objs, tx)
	return objs, allErr
}
