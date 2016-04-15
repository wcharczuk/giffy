package spiffy

import (
	"database/sql"
	"testing"

	"github.com/blendlabs/go-assert"
)

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

// TestConnectionSanityCheck tests if we can connect to the db, a.k.a., if the underlying driver works.
func TestConnectionSanityCheck(t *testing.T) {
	config := dbConnectionFromEnvironment()
	_, dbErr := sql.Open("postgres", config.CreatePostgresConnectionString())
	if dbErr != nil {
		t.Error("Error opening database")
		t.FailNow()
	}
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

	objs := []benchObj{}
	queryErr := DefaultDb().QueryInTransaction("select * from bench_object", tx).OutMany(&objs)

	a.Nil(queryErr)
	a.NotEmpty(objs)

	all := []benchObj{}
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.Equal(len(objs), len(all))

	obj := benchObj{}
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
	defer tx.Rollback()

	seedErr := seedObjects(100, tx)
	a.Nil(seedErr)

	objs := []benchObj{}
	queryErr := DefaultDb().QueryInTransaction("select * from bench_object", tx).OutMany(&objs)

	a.Nil(queryErr)
	a.NotEmpty(objs)

	all := []benchObj{}
	allErr := DefaultDb().GetAllInTransaction(&all, tx)
	a.Nil(allErr)
	a.Equal(len(objs), len(all))

	sampleObj := all[0]

	getTest := benchObj{}
	getTestErr := DefaultDb().GetByIDInTransaction(&getTest, tx, sampleObj.ID)
	a.Nil(getTestErr)
	a.Equal(sampleObj.ID, getTest.ID)

	exists, existsErr := DefaultDb().ExistsInTransaction(&getTest, tx)
	a.Nil(existsErr)
	a.True(exists)

	getTest.Name = "not_a_test_object"

	updateErr := DefaultDb().UpdateInTransaction(&getTest, tx)
	a.Nil(updateErr)

	verify := benchObj{}
	verifyErr := DefaultDb().GetByIDInTransaction(&verify, tx, getTest.ID)
	a.Nil(verifyErr)
	a.Equal(getTest.Name, verify.Name)

	deleteErr := DefaultDb().DeleteInTransaction(&verify, tx)
	a.Nil(deleteErr)

	delVerify := benchObj{}
	delVerifyErr := DefaultDb().GetByIDInTransaction(&delVerify, tx, getTest.ID)
	a.Nil(delVerifyErr)
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

func TestIsolateToTransaction(t *testing.T) {
	a := assert.New(t)

	tx, err := DefaultDb().Begin()
	a.Nil(err)

	DefaultDb().IsolateToTransaction(tx)
	defer DefaultDb().ReleaseIsolation()
	a.NotNil(DefaultDb().Tx)
	a.True(DefaultDb().IsIsolatedToTransaction())
}

func TestReleaseIsolation(t *testing.T) {
	a := assert.New(t)

	tx, err := DefaultDb().Begin()
	a.Nil(err)

	DefaultDb().IsolateToTransaction(tx)
	defer DefaultDb().ReleaseIsolation() //this has to happen regardless (panics etc.)

	a.NotNil(DefaultDb().Tx)
	a.True(DefaultDb().IsIsolatedToTransaction())

	DefaultDb().ReleaseIsolation()
	a.Nil(DefaultDb().Tx)
	a.False(DefaultDb().IsIsolatedToTransaction())
}

func TestBeginReturnsIsolatedTransaction(t *testing.T) {
	a := assert.New(t)

	tx, err := DefaultDb().Begin()
	a.Nil(err)

	DefaultDb().IsolateToTransaction(tx)
	defer DefaultDb().ReleaseIsolation()

	currentTx, err := DefaultDb().Begin()
	a.Nil(err)
	a.Equal(tx, currentTx)
}
