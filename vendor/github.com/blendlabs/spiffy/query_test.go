package spiffy

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestQueryExecute(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	stmt, rows, err := Default().QueryInTx("select * from bench_object", tx).Execute()
	a.Nil(err)
	defer stmt.Close()
	defer rows.Close()
	a.NotNil(rows)
	a.NotNil(stmt)
}

func TestQueryEach(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []benchObj
	var popErr error
	err = Default().QueryInTx("select * from bench_object", tx).Each(func(r *sql.Rows) error {
		bo := benchObj{}
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

func TestQueryAny(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	err = seedObjects(10, tx)
	a.Nil(err)

	var all []benchObj
	allErr := Default().GetAllInTx(&all, tx)
	a.Nil(allErr)
	a.NotEmpty(all)

	obj := all[0]

	exists, err := Default().QueryInTx("select 1 from bench_object where id = $1", tx, obj.ID).Any()
	a.Nil(err)
	a.True(exists)

	notExists, err := Default().QueryInTx("select 1 from bench_object where id = $1", tx, -1).Any()
	a.Nil(err)
	a.False(notExists)
}

func TestQueryNone(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []benchObj
	allErr := Default().GetAllInTx(&all, tx)
	a.Nil(allErr)
	a.NotEmpty(all)

	obj := all[0]

	exists, existsErr := Default().QueryInTx("select 1 from bench_object where id = $1", tx, obj.ID).None()
	a.Nil(existsErr)
	a.False(exists)

	notExists, notExistsErr := Default().QueryInTx("select 1 from bench_object where id = $1", tx, -1).None()
	a.Nil(notExistsErr)
	a.True(notExists)
}

func TestQueryPanicHandling(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	err = seedObjects(10, tx)
	a.Nil(err)

	err = Default().QueryInTx("select * from bench_object", tx).Each(func(r *sql.Rows) error {
		panic("THIS IS A TEST PANIC")
	})
	a.NotNil(err) // this should have the result of the panic

	// we now test to see if the connection is still in a good state, i.e. that we recovered from the panic
	// and closed the connection / rows / statement
	hasRows, err := Default().QueryInTx("select * from bench_object", tx).Any()
	a.Nil(err)
	a.True(hasRows)
}

func TestMultipleQueriesPerTransaction(t *testing.T) {
	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	wg := sync.WaitGroup{}
	wg.Add(3)

	a.NotNil(Default().Connection)

	err = seedObjects(10, tx)
	a.Nil(err)

	go func() {
		defer wg.Done()
		hasRows, err := Default().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := Default().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := Default().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	wg.Wait()

	hasRows, err := Default().Query("select * from bench_object").Any()
	a.Nil(err)
	a.True(hasRows)
}

// Note: this test assumes that `bench_object` DOES NOT EXIST.
// It also is generally skipped as it barfs a bunch of errors into the
// postgres log.
func TestMultipleQueriesPerTransactionWithFailure(t *testing.T) {
	t.Skip()

	a := assert.New(t)
	tx, err := Default().Begin()
	a.Nil(err)
	defer tx.Rollback()

	wg := sync.WaitGroup{}
	wg.Add(3)

	a.NotNil(Default().Connection)

	go func() {
		defer wg.Done()
		hasRows, err := Default().QueryInTx("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := Default().QueryInTx("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := Default().QueryInTx("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	wg.Wait()
	hasRows, err := Default().QueryInTx("select * from bench_object", tx).Any()

	a.NotNil(err)
	a.False(hasRows)
}
