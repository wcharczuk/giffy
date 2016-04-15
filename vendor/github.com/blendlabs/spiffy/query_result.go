package spiffy

import (
	"database/sql"
	"reflect"

	"github.com/blendlabs/go-exception"
)

// --------------------------------------------------------------------------------
// Query Result
// --------------------------------------------------------------------------------

// QueryResult is the intermediate result of a query.
type QueryResult struct {
	Rows  *sql.Rows
	Stmt  *sql.Stmt
	Conn  *DbConnection
	Error error
}

// Close closes and releases any resources retained by the QueryResult.
func (q *QueryResult) Close() error {
	var rowsErr error
	var stmtErr error

	if q.Rows != nil {
		rowsErr = q.Rows.Close()
		q.Rows = nil
	}
	if q.Stmt != nil {
		stmtErr = q.Stmt.Close()
		q.Stmt = nil
	}

	//yes this is gross.
	//release the tx lock on the connection for this query.
	q.Conn.txUnlock()
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
	}()

	if q.Error != nil {
		hasRows = false
		err = exception.Wrap(q.Error)
		return
	}

	rowsErr := q.Rows.Err()
	if rowsErr != nil {
		hasRows = false
		err = exception.Wrap(rowsErr)
		return
	}

	hasRows = q.Rows.Next()
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

	if q.Error != nil {
		hasRows = false
		err = exception.Wrap(q.Error)
		return
	}

	rowsErr := q.Rows.Err()
	if rowsErr != nil {
		hasRows = false
		err = exception.Wrap(rowsErr)
		return
	}

	hasRows = !q.Rows.Next()
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
	}()

	if q.Error != nil {
		err = exception.Wrap(q.Error)
		return
	}

	rowsErr := q.Rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	if q.Rows.Next() {
		scanErr := q.Rows.Scan(args...)
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
	}()

	if q.Error != nil {
		err = exception.Wrap(q.Error)
		return
	}

	rowsErr := q.Rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	columnMeta := CachedColumnCollectionFromInstance(object)
	var popErr error
	if q.Rows.Next() {
		if populatable, isPopulatable := object.(Populatable); isPopulatable {
			popErr = populatable.Populate(q.Rows)
		} else {
			popErr = PopulateByName(object, q.Rows, columnMeta)
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
	}()

	if q.Error != nil {
		err = exception.Wrap(q.Error)
		return err
	}

	rowsErr := q.Rows.Err()
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
	meta := CachedColumnCollectionFromType(v.TableName(), sliceInnerType)

	isPopulatable := IsPopulatable(v)

	var popErr error
	didSetRows := false
	for q.Rows.Next() {
		newObj, _ := MakeNew(sliceInnerType)

		if isPopulatable {
			popErr = AsPopulatable(newObj).Populate(q.Rows)
		} else {
			popErr = PopulateByName(newObj, q.Rows, meta)
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
	}()

	if q.Error != nil {
		return q.Error
	}

	rowsErr := q.Rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	for q.Rows.Next() {
		err = consumer(q.Rows)
		if err != nil {
			return err
		}
	}
	return
}
