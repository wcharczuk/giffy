package spiffy

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-exception"
)

//------------------------------------------------------------------------------------------------
// Testing Entrypoint
//------------------------------------------------------------------------------------------------

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

	return &DbConnection{Host: dbHost, Schema: dbSchema, Username: dbUser, Password: dbPassword, SSLMode: "disable"}
}

//------------------------------------------------------------------------------------------------
// Benchmarking
//------------------------------------------------------------------------------------------------

type BenchObj struct {
	ID        int       `db:"id,pk,serial"`
	Name      string    `db:"name"`
	Timestamp time.Time `db:"timestamp_utc"`
	Amount    float32   `db:"amount"`
	Pending   bool      `db:"pending"`
	Category  string    `db:"category"`
}

func (b *BenchObj) Populate(rows *sql.Rows) error {
	var id int
	var name string
	var ts time.Time
	var amount float32
	var pending bool
	var category string
	scanErr := rows.Scan(&id, &name, &ts, &amount, &pending, &category)

	if scanErr != nil {
		return scanErr
	}

	b.ID = id
	b.Name = name
	b.Timestamp = ts
	b.Amount = amount
	b.Pending = pending
	b.Category = category
	return nil
}

func (b BenchObj) TableName() string {
	return "bench_object"
}

func createTable(tx *sql.Tx) error {
	createSQL := `CREATE TABLE bench_object (id serial not null, name varchar(255), timestamp_utc timestamp, amount real, pending boolean, category varchar(255));`
	return DefaultDb().ExecInTransaction(createSQL, tx)
}

func createObject(index int, tx *sql.Tx) error {
	obj := BenchObj{}
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

func readManual(tx *sql.Tx) ([]BenchObj, error) {
	objs := []BenchObj{}
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
		obj := &BenchObj{}
		populateErr := obj.Populate(rows)
		if populateErr != nil {
			return nil, populateErr
		}
		objs = append(objs, *obj)
	}

	return objs, nil
}

func readOrm(tx *sql.Tx) ([]BenchObj, error) {
	objs := []BenchObj{}
	allErr := DefaultDb().GetAllInTransaction(&objs, tx)
	return objs, allErr
}

func BenchmarkMain(b *testing.B) {
	tx, txErr := DefaultDb().Begin()
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
// Testing
//------------------------------------------------------------------------------------------------

func TestNewAunauthenticatedDbConnection(t *testing.T) {
	a := assert.New(t)
	conn := NewUnauthenticatedDbConnection("test_host", "test_schema")
	a.Equal("test_host", conn.Host)
	a.Equal("test_schema", conn.Schema)
}

func TestNewDbConnection(t *testing.T) {
	a := assert.New(t)
	conn := NewDbConnection("test_host", "test_schema", "test_user", "test_password")
	a.Equal("test_host", conn.Host)
	a.Equal("test_schema", conn.Schema)
	a.Equal("test_user", conn.Username)
	a.Equal("test_password", conn.Password)
}

func TestNewSSLDbConnection(t *testing.T) {
	a := assert.New(t)
	conn := NewSSLDbConnection("test_host", "test_schema", "test_user", "test_password", "a good one")
	a.Equal("test_host", conn.Host)
	a.Equal("test_schema", conn.Schema)
	a.Equal("test_user", conn.Username)
	a.Equal("test_password", conn.Password)
	a.Equal("a good one", conn.SSLMode)
}

func TestSanityCheck(t *testing.T) {
	config := dbConnectionFromEnvironment()
	_, dbErr := sql.Open("postgres", config.CreatePostgresConnectionString())
	if dbErr != nil {
		t.Error("Error opening database")
		t.FailNow()
	}
}

func TestAliases(t *testing.T) {
	oldDefaultAlias := defaultAlias
	defaultAlias = ""
	defer func() {
		defaultAlias = oldDefaultAlias
	}()

	a := assert.New(t)
	config := dbConnectionFromEnvironment()

	CreateDbAlias("test", config)

	gotConn := Alias("test")
	a.Equal(config.Username, gotConn.Username, "Alias(name) should return the correct alias.")

	shouldBeNil := DefaultDb()
	a.Nil(shouldBeNil, "DefaultDb() without an alias should return nil.", fmt.Sprintf("%v", shouldBeNil))

	SetDefaultAlias("test")
	defaultConn := DefaultDb()
	a.NotNil(defaultConn, "DefaultDb() with an alias should return the aliased connection.")
}

func TestTransactionIsolation(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer a.Nil(tx.Rollback())

	DefaultDb().IsolateToTransaction(tx)
	a.True(DefaultDb().Tx != nil)
	a.True(DefaultDb().IsIsolatedToTransaction())

	_, tx2Err := DefaultDb().Begin()
	a.Nil(tx2Err)

	DefaultDb().ReleaseIsolation()
	a.False(DefaultDb().IsIsolatedToTransaction())

}

func TestPrepare(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	createTableEsrr := createTable(tx)
	a.Nil(createTableEsrr)
	a.Nil(tx.Rollback())
}

func TestQuery(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	seedErr := seedObjects(100, tx)
	a.Nil(seedErr)

	objs := []BenchObj{}
	queryErr := DefaultDb().QueryInTransaction("select * from bench_object", tx).OutMany(&objs)

	a.Nil(queryErr)
	a.NotEmpty(objs)

	all := []BenchObj{}
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.Equal(len(objs), len(all))

	obj := BenchObj{}
	singleQueryErr := DefaultDb().QueryInTransaction("select * from bench_object limit 1", tx).Out(&obj)
	a.Nil(singleQueryErr)
	a.NotEqual(obj.ID, 0)

	var id int
	scanErr := DefaultDb().QueryInTransaction("select id from bench_object limit 1", tx).Scan(&id)
	a.Nil(scanErr)
	a.NotEqual(id, 0)
}

func TestCRUDMethods(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	seedErr := seedObjects(100, tx)
	a.Nil(seedErr)

	objs := []BenchObj{}
	queryErr := DefaultDb().QueryInTransaction("select * from bench_object", tx).OutMany(&objs)

	a.Nil(queryErr)
	a.NotEmpty(objs)

	all := []BenchObj{}
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.Equal(len(objs), len(all))

	sampleObj := all[0]

	getTest := BenchObj{}
	getTestErr := DefaultDb().GetByIDInTransaction(&getTest, tx, sampleObj.ID)
	a.Nil(getTestErr)
	a.Equal(sampleObj.ID, getTest.ID)

	exists, existsErr := DefaultDb().ExistsInTransaction(&getTest, tx)
	a.Nil(existsErr)
	a.True(exists)

	getTest.Name = "not_a_test_object"

	updateErr := DefaultDb().UpdateInTransaction(&getTest, tx)
	a.Nil(updateErr)

	verify := BenchObj{}
	verifyErr := DefaultDb().GetByIDInTransaction(&verify, tx, getTest.ID)
	a.Nil(verifyErr)
	a.Equal(getTest.Name, verify.Name)

	deleteErr := DefaultDb().DeleteInTransaction(&verify, tx)
	a.Nil(deleteErr)

	delVerify := BenchObj{}
	delVerifyErr := DefaultDb().GetByIDInTransaction(&delVerify, tx, getTest.ID)
	a.Nil(delVerifyErr)
}

type myStruct struct {
	PrimaryKeyCol     int    `json:"pk" db:"primary_key_column,pk,serial"`
	InferredName      string `json:"normal"`
	Excluded          string `json:"-" db:"-"`
	NullableCol       string `json:"not_nullable" db:"nullable,nullable"`
	InferredWithFlags string `db:",readonly"`
}

func (m myStruct) TableName() string {
	return "my_struct"
}

func TestGetColumns(t *testing.T) {
	a := assert.New(t)

	emptyColumnCollection := ColumnCollection{}
	firstOrDefaultNil := emptyColumnCollection.FirstOrDefault()
	a.Nil(firstOrDefaultNil)

	obj := myStruct{}
	meta := NewColumnCollectionFromInstance(obj)

	a.NotNil(meta.Columns)
	a.NotEmpty(meta.Columns)

	a.Equal(4, len(meta.Columns))

	readOnlyColumns := meta.ReadOnly()
	a.Len(readOnlyColumns.Columns, 1)

	firstOrDefault := meta.FirstOrDefault()
	a.NotNil(firstOrDefault)

	firstCol := meta.Columns[0]
	a.Equal("my_struct", firstCol.TableName)
	a.Equal("PrimaryKeyCol", firstCol.FieldName)
	a.Equal("primary_key_column", firstCol.ColumnName)
	a.True(firstCol.IsPrimaryKey)
	a.True(firstCol.IsSerial)
	a.False(firstCol.IsNullable)
	a.False(firstCol.IsReadOnly)

	secondCol := meta.Columns[1]
	a.Equal("inferredname", secondCol.ColumnName)
	a.False(secondCol.IsPrimaryKey)
	a.False(secondCol.IsSerial)
	a.False(secondCol.IsNullable)
	a.False(secondCol.IsReadOnly)

	thirdCol := meta.Columns[2]
	a.Equal("nullable", thirdCol.ColumnName)
	a.False(thirdCol.IsPrimaryKey)
	a.False(thirdCol.IsSerial)
	a.True(thirdCol.IsNullable)
	a.False(thirdCol.IsReadOnly)

	fourthCol := meta.Columns[3]
	a.Equal("inferredwithflags", fourthCol.ColumnName)
	a.False(fourthCol.IsPrimaryKey)
	a.False(fourthCol.IsSerial)
	a.False(fourthCol.IsNullable)
	a.True(fourthCol.IsReadOnly)
}

func TestSetValue(t *testing.T) {
	a := assert.New(t)
	obj := myStruct{InferredName: "Hello."}

	var value interface{}
	value = 10
	meta := NewColumnCollectionFromInstance(obj)
	pk := meta.Columns[0]
	a.Nil(pk.SetValue(&obj, value))
	a.Equal(10, obj.PrimaryKeyCol)
}

func TestGetValue(t *testing.T) {
	a := assert.New(t)
	obj := myStruct{PrimaryKeyCol: 5, InferredName: "Hello."}

	meta := NewColumnCollectionFromInstance(obj)
	pk := meta.PrimaryKeys().Columns[0]
	value := pk.GetValue(&obj)
	a.NotNil(value)
	a.Equal(5, value)
}

func TestMakeCsvTokens(t *testing.T) {
	a := assert.New(t)

	one := makeCsvTokens(1)
	two := makeCsvTokens(2)
	three := makeCsvTokens(3)

	a.Equal("$1", one)
	a.Equal("$1,$2", two)
	a.Equal("$1,$2,$3", three)
}

func TestMakeSliceOfType(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	myType := reflectType(BenchObj{})
	sliceOfT, castOk := makeSliceOfType(myType).(*[]BenchObj)
	a.True(castOk)

	allErr := DefaultDb().GetAllInTransaction(sliceOfT, tx)
	a.Nil(allErr)
	a.NotEmpty(*sliceOfT)
}

func TestDbConnectionOpen(t *testing.T) {
	a := assert.New(t)

	testAlias := dbConnectionFromEnvironment()
	db, dbErr := testAlias.Open()
	a.Nil(dbErr)
	a.NotNil(db)
	defer db.Close()
}

func TestExec(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	execErr := DefaultDb().ExecInTransaction("select 'ok!'", tx)
	a.Nil(execErr)
}

func TestQueryResultEach(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []BenchObj
	var popErr error
	err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Each(func(r *sql.Rows) error {
		bo := BenchObj{}
		popErr = bo.Populate(r)
		if popErr != nil {
			return popErr
		}
		all = append(all, bo)
		return nil
	})
	a.Nil(err)
	a.NotEmpty(all)
}

func TestQueryResultAny(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []BenchObj
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.NotEmpty(all)

	obj := all[0]

	exists, existsErr := DefaultDb().QueryInTransaction("select 1 from bench_object where id = $1", tx, obj.ID).Any()
	a.Nil(existsErr)
	a.True(exists)

	notExists, notExistsErr := DefaultDb().QueryInTransaction("select 1 from bench_object where id = $1", tx, -1).Any()
	a.Nil(notExistsErr)
	a.False(notExists)
}

func TestQueryResultNone(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []BenchObj
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.NotEmpty(all)

	obj := all[0]

	exists, existsErr := DefaultDb().QueryInTransaction("select 1 from bench_object where id = $1", tx, obj.ID).None()
	a.Nil(existsErr)
	a.False(exists)

	notExists, notExistsErr := DefaultDb().QueryInTransaction("select 1 from bench_object where id = $1", tx, -1).None()
	a.Nil(notExistsErr)
	a.True(notExists)
}
