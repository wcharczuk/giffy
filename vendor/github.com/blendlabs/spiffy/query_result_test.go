package spiffy

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestQueryResultEach(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []benchObj
	var popErr error
	err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Each(func(r *sql.Rows) error {
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

func TestQueryResultAny(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer tx.Rollback()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	var all []benchObj
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

	var all []benchObj
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

func TestQueryResultPanicHandling(t *testing.T) {
	a := assert.New(t)
	tx, err := DefaultDb().Begin()
	a.Nil(err)
	defer tx.Rollback()

	err = seedObjects(10, tx)
	a.Nil(err)

	err = DefaultDb().QueryInTransaction("select * from bench_object", tx).Each(func(r *sql.Rows) error {
		panic("THIS IS A TEST PANIC")
	})
	a.NotNil(err) // this should have the result of the panic

	// we now test to see if the connection is still in a good state, i.e. that we recovered from the panic
	// and closed the connection / rows / statement
	hasRows, err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Any()
	a.Nil(err)
	a.True(hasRows)
}

func TestMultipleQueriesPerTransaction(t *testing.T) {
	a := assert.New(t)
	tx, err := DefaultDb().Begin()
	a.Nil(err)
	defer tx.Rollback()

	DefaultDb().IsolateToTransaction(tx)
	defer DefaultDb().ReleaseIsolation()

	wg := sync.WaitGroup{}
	wg.Add(3)

	a.NotNil(DefaultDb().Connection)
	a.NotNil(DefaultDb().Tx)

	err = seedObjects(10, tx)
	a.Nil(err)

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().Query("select * from bench_object").Any()
		a.Nil(err)
		a.True(hasRows)
	}()

	wg.Wait()

	hasRows, err := DefaultDb().Query("select * from bench_object").Any()
	a.Nil(err)
	a.True(hasRows)
}

// Note: this test assumes that `bench_object` DOES NOT EXIST.
func TestMultipleQueriesPerTransactionWithFailure(t *testing.T) {
	a := assert.New(t)
	tx, err := DefaultDb().Begin()
	a.Nil(err)
	defer tx.Rollback()

	wg := sync.WaitGroup{}
	wg.Add(3)

	a.NotNil(DefaultDb().Connection)

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	go func() {
		defer wg.Done()
		hasRows, err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Any()
		a.NotNil(err)
		a.False(hasRows)
	}()

	wg.Wait()
	hasRows, err := DefaultDb().QueryInTransaction("select * from bench_object", tx).Any()

	a.NotNil(err)
	a.False(hasRows)
}
