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
	err := OpenDefault(NewFromEnv())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// BenchmarkMain is the benchmarking entrypoint.
func BenchmarkMain(b *testing.B) {
	tx, txErr := Default().Begin()
	if txErr != nil {
		b.Error("Unable to create transaction")
		b.FailNow()
	}
	if tx == nil {
		b.Error("`tx` is nil")
		b.FailNow()
	}

	defer func() {
		if tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				b.Errorf("Error rolling back transaction: %v", rollbackErr)
				b.FailNow()
			}
		}
	}()

	seedErr := seedObjects(5000, tx)
	if seedErr != nil {
		b.Errorf("Error seeding objects: %v", seedErr)
		b.FailNow()
	}

	manualBefore := time.Now()
	_, manualErr := readManual(tx)
	manualAfter := time.Now()
	if manualErr != nil {
		b.Errorf("Error using manual query: %v", manualErr)
		b.FailNow()
	}

	ormBefore := time.Now()
	_, ormErr := readOrm(tx)
	ormAfter := time.Now()
	if ormErr != nil {
		b.Errorf("Error using orm: %v", ormErr)
		b.FailNow()
	}

	b.Logf("Benchmark Test Results: Manual: %v vs. Orm: %v\n", manualAfter.Sub(manualBefore), ormAfter.Sub(ormBefore))
}

//------------------------------------------------------------------------------------------------
// Util Types
//------------------------------------------------------------------------------------------------

type upsertObj struct {
	UUID      string    `db:"uuid,pk"`
	Timestamp time.Time `db:"timestamp_utc"`
	Category  string    `db:"category"`
}

func (uo upsertObj) TableName() string {
	return "upsert_object"
}

func createUpserObjectTable(tx *sql.Tx) error {
	createSQL := `CREATE TABLE IF NOT EXISTS upsert_object (uuid varchar(255) primary key, timestamp_utc timestamp, category varchar(255));`
	return Default().ExecInTx(createSQL, tx)
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
	createSQL := `CREATE TABLE IF NOT EXISTS bench_object (id serial not null primary key, name varchar(255), timestamp_utc timestamp, amount real, pending boolean, category varchar(255));`
	return Default().ExecInTx(createSQL, tx)
}

func dropTable(tx *sql.Tx) error {
	dropSQL := `DROP TABLE IF NOT EXISTS bench_object;`
	return Default().ExecInTx(dropSQL, tx)
}

func createObject(index int, tx *sql.Tx) error {
	obj := benchObj{}
	obj.Name = fmt.Sprintf("test_object_%d", index)
	obj.Timestamp = time.Now().UTC()
	obj.Amount = 1000.0 + (5.0 * float32(index))
	obj.Pending = index%2 == 0
	obj.Category = fmt.Sprintf("category_%d", index)
	return exception.Wrap(Default().CreateInTx(&obj, tx))
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
	readStmt, readStmtErr := Default().Prepare(readSQL, tx)
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
	var objs []benchObj
	allErr := Default().GetAllInTx(&objs, tx)
	return objs, allErr
}
