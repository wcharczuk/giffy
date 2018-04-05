package db

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// --------------------------------------------------------------------------------
// Query Result
// --------------------------------------------------------------------------------

// Query is the intermediate result of a query.
type Query struct {
	statement      string
	statementLabel string
	args           []interface{}

	start time.Time
	rows  *sql.Rows

	stmt       *sql.Stmt
	fireEvents bool
	conn       *Connection
	ctx        context.Context
	tx         *sql.Tx
	err        error
}

// Close closes and releases any resources retained by the QueryResult.
func (q *Query) Close() error {
	var rowsErr error
	var stmtErr error

	if q.rows != nil {
		rowsErr = q.rows.Close()
		q.rows = nil
	}

	if !q.conn.useStatementCache {
		if q.stmt != nil {
			stmtErr = q.stmt.Close()
			q.stmt = nil
		}
	}
	return exception.Nest(rowsErr, stmtErr)
}

// CachedAs sets the statement cache label for the query.
func (q *Query) CachedAs(cacheLabel string) *Query {
	q.statementLabel = cacheLabel
	return q
}

// Execute runs a given query, yielding the raw results.
func (q *Query) Execute() (stmt *sql.Stmt, rows *sql.Rows, err error) {
	var stmtErr error
	if q.shouldCacheStatement() {
		stmt, stmtErr = q.conn.PrepareCached(q.statementLabel, q.statement, q.tx)
	} else {
		stmt, stmtErr = q.conn.Prepare(q.statement, q.tx)
	}

	if stmtErr != nil {
		if q.shouldCacheStatement() {
			q.conn.statementCache.InvalidateStatement(q.statementLabel)
		}
		err = exception.Wrap(stmtErr)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if q.conn.useStatementCache {
				err = exception.Nest(err, exception.New(r))
			} else {
				err = exception.Nest(err, exception.New(r), stmt.Close())
			}
		}
	}()

	var queryErr error
	if q.ctx != nil {
		rows, queryErr = stmt.QueryContext(q.ctx, q.args...)
	} else {
		rows, queryErr = stmt.Query(q.args...)
	}

	if queryErr != nil {
		if q.shouldCacheStatement() {
			q.conn.statementCache.InvalidateStatement(q.statementLabel)
		}
		err = exception.Wrap(queryErr)
	}
	return
}

// Any returns if there are any results for the query.
func (q *Query) Any() (hasRows bool, err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()
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
func (q *Query) None() (hasRows bool, err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()

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
func (q *Query) Scan(args ...interface{}) (err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()
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
func (q *Query) Out(object interface{}) (err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()
	if q.err != nil {
		err = exception.Wrap(q.err)
		return
	}

	rowsErr := q.rows.Err()
	if rowsErr != nil {
		err = exception.Wrap(rowsErr)
		return
	}

	sliceType := reflectType(object)
	if sliceType.Kind() != reflect.Struct {
		err = exception.New("destination object is not a struct")
		return
	}

	columnMeta := getCachedColumnCollectionFromInstance(object)
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
func (q *Query) OutMany(collection interface{}) (err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()
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
		err = exception.New("destination collection is not a slice")
		return
	}

	sliceInnerType := reflectSliceType(collection)
	collectionValue := reflectValue(collection)

	v := makeNew(sliceInnerType)
	meta := getCachedColumnCollectionFromType(newColumnCacheKey(sliceInnerType), sliceInnerType)

	isPopulatable := isPopulatable(v)

	var popErr error
	didSetRows := false
	for q.rows.Next() {
		newObj := makeNew(sliceInnerType)

		if isPopulatable {
			popErr = asPopulatable(newObj).Populate(q.rows)
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
func (q *Query) Each(consumer RowsConsumer) (err error) {
	defer func() { err = q.finalizer(recover(), err) }()

	q.stmt, q.rows, q.err = q.Execute()
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

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

func (q *Query) finalizer(r interface{}, err error) error {
	if r != nil {
		recoveryException := exception.New(r)
		err = exception.Nest(err, recoveryException)
	}

	if closeErr := q.Close(); closeErr != nil {
		err = exception.Nest(err, closeErr)
	}

	if q.fireEvents {
		q.conn.fireEvent(logger.Query, q.statement, time.Since(q.start), err, q.statementLabel)
	}
	return err
}

func (q *Query) shouldCacheStatement() bool {
	return q.conn.useStatementCache && len(q.statementLabel) > 0
}
