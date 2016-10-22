package spiffy

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/blendlabs/go-exception"
)

// --------------------------------------------------------------------------------
// Query Result
// --------------------------------------------------------------------------------

// QueryResult is the intermediate result of a query.
type QueryResult struct {
	start     time.Time
	rows      *sql.Rows
	queryBody string
	stmt      *sql.Stmt
	conn      *DbConnection
	err       error
}

// Close closes and releases any resources retained by the QueryResult.
func (q *QueryResult) Close() error {
	var rowsErr error
	var stmtErr error

	if q.rows != nil {
		rowsErr = q.rows.Close()
		q.rows = nil
	}
	if q.stmt != nil {
		stmtErr = q.stmt.Close()
		q.stmt = nil
	}

	//yes this is gross.
	//release the tx lock on the connection for this query.
	q.conn.txUnlock()
	return exception.WrapMany(rowsErr, stmtErr)
}

// Any returns if there are any results for the query.
func (q *QueryResult) Any() (hasRows bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
		q.conn.FireEvent(q.conn.queryListeners, q.queryBody, time.Now().Sub(q.start), err)
	}()

	if q.err != nil {
		hasRows = false
		err = exception.Wrap(q.err)
		return
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		hasRows = false
		err = exception.Wrap(rowsErr)
		return
	}

	hasRows = q.rows.Next()
	return
}

// None returns if there are no results for the query.
func (q *QueryResult) None() (hasRows bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	if q.err != nil {
		hasRows = false
		err = exception.Wrap(q.err)
		return
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		hasRows = false
		err = exception.Wrap(rowsErr)
		return
	}

	hasRows = !q.rows.Next()
	return
}

// Scan writes the results to a given set of local variables.
func (q *QueryResult) Scan(args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
		q.conn.FireEvent(q.conn.queryListeners, q.queryBody, time.Now().Sub(q.start), err)
	}()

	if q.err != nil {
		err = exception.Wrap(q.err)
		return
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	if q.rows.Next() {
		scanErr := q.rows.Scan(args...)
		if scanErr != nil {
			err = exception.Wrap(scanErr)
		}
	}

	return
}

// Out writes the query result to a single object via. reflection mapping.
func (q *QueryResult) Out(object DatabaseMapped) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
		q.conn.FireEvent(q.conn.queryListeners, q.queryBody, time.Now().Sub(q.start), err)
	}()

	if q.err != nil {
		err = exception.Wrap(q.err)
		return
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	columnMeta := CachedColumnCollectionFromInstance(object)
	var popErr error
	if q.rows.Next() {
		if populatable, isPopulatable := object.(Populatable); isPopulatable {
			popErr = populatable.Populate(q.rows)
		} else {
			popErr = PopulateByName(object, q.rows, columnMeta)
		}
		if popErr != nil {
			err = popErr
			return
		}
	}

	return
}

// OutMany writes the query results to a slice of objects.
func (q *QueryResult) OutMany(collection interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
		q.conn.FireEvent(q.conn.queryListeners, q.queryBody, time.Now().Sub(q.start), err)
	}()

	if q.err != nil {
		err = exception.Wrap(q.err)
		return err
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	sliceType := reflectType(collection)
	if sliceType.Kind() != reflect.Slice {
		err = exception.New("Destination collection is not a slice.")
		return
	}

	sliceInnerType := reflectSliceType(collection)
	collectionValue := reflectValue(collection)

	v, _ := MakeNew(sliceInnerType)
	meta := CachedColumnCollectionFromType(MakeColumnCacheKey(sliceInnerType, v.TableName()), sliceInnerType)

	isPopulatable := IsPopulatable(v)

	var popErr error
	didSetRows := false
	for q.rows.Next() {
		newObj, _ := MakeNew(sliceInnerType)

		if isPopulatable {
			popErr = AsPopulatable(newObj).Populate(q.rows)
		} else {
			popErr = PopulateByName(newObj, q.rows, meta)
		}

		if popErr != nil {
			err = popErr
			return
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
		didSetRows = true
	}

	if !didSetRows {
		collectionValue.Set(reflect.MakeSlice(sliceType, 0, 0))
	}
	return
}

// Each writes the query results to a slice of objects.
func (q *QueryResult) Each(consumer RowsConsumer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}

		if closeErr := q.Close(); closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
		q.conn.FireEvent(q.conn.queryListeners, q.queryBody, time.Now().Sub(q.start), err)
	}()

	if q.err != nil {
		return q.err
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	for q.rows.Next() {
		err = consumer(q.rows)
		if err != nil {
			return err
		}
	}
	return
}
