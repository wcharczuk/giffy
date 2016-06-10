// Package spiffy providers a basic abstraction layer above normal database/sql that makes it easier to
// interact with the database and organize database related code. It is not intended to replace actual sql
// (you write queries yourself).
package spiffy

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/blendlabs/go-exception"

	// PQ is the postgres driver
	_ "github.com/lib/pq"
)

const (
	//DBAliasNilError is a common error
	DBAliasNilError = "DbConnection is nil; did you set up a DbAlias for your project?"
)

// --------------------------------------------------------------------------------
// DbConnection
// --------------------------------------------------------------------------------

// NewUnauthenticatedDbConnection creates a new DbConnection without Username or Password or SSLMode
func NewUnauthenticatedDbConnection(host, schema string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = ""
	conn.Password = ""
	conn.SSLMode = "disable"
	conn.MetaLock = &sync.Mutex{}
	conn.TxLock = &sync.RWMutex{}
	return conn
}

// NewDbConnection creates a new connection with SSLMode set to "disable"
func NewDbConnection(host, schema, username, password string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = username
	conn.Password = password
	conn.SSLMode = "disable"
	conn.MetaLock = &sync.Mutex{}
	conn.TxLock = &sync.RWMutex{}
	return conn
}

// NewDbConnectionFromDSN creates a new connection with SSLMode set to "disable"
func NewDbConnectionFromDSN(dsn string) *DbConnection {
	conn := &DbConnection{}
	conn.DSN = dsn
	conn.MetaLock = &sync.Mutex{}
	conn.TxLock = &sync.RWMutex{}
	return conn
}

// NewSSLDbConnection creates a new connection with all available options (including SSLMode)
func NewSSLDbConnection(host, schema, username, password, sslMode string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = username
	conn.Password = password
	conn.SSLMode = sslMode
	conn.MetaLock = &sync.Mutex{}
	conn.TxLock = &sync.RWMutex{}
	return conn
}

// DbConnection is the basic wrapper for connection parameters and saves a reference to the created sql.Connection.
type DbConnection struct {
	Host       string
	Schema     string
	Username   string
	Password   string
	SSLMode    string
	DSN        string
	Connection *sql.DB
	MetaLock   *sync.Mutex

	Tx     *sql.Tx
	TxLock *sync.RWMutex
}

// CreatePostgresConnectionString returns a sql connection string from a given set of DbConnection parameters.
func (dbc *DbConnection) CreatePostgresConnectionString() string {
	if len(dbc.DSN) != 0 {
		return dbc.DSN
	}
	sslMode := "?sslmode=disable"
	if dbc.SSLMode != "" {
		sslMode = fmt.Sprintf("?sslmode=%s", dbc.SSLMode)
	}

	if dbc.Username != "" {
		if dbc.Password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s/%s%s", dbc.Username, dbc.Password, dbc.Host, dbc.Schema, sslMode)
		}
		return fmt.Sprintf("postgres://%s@%s/%s%s", dbc.Username, dbc.Host, dbc.Schema, sslMode)
	}
	return fmt.Sprintf("postgres://%s/%s%s", dbc.Host, dbc.Schema, sslMode)
}

// Begin starts a new transaction.
func (dbc *DbConnection) Begin() (*sql.Tx, error) {
	if dbc == nil {
		return nil, exception.New(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return dbc.Tx, nil
	}

	if dbc.Connection != nil {
		tx, txErr := dbc.Connection.Begin()
		return tx, exception.Wrap(txErr)
	}

	connection, err := dbc.Open()
	if err != nil {
		return nil, exception.Wrap(err)
	}
	tx, err := connection.Begin()
	return tx, exception.Wrap(err)
}

// WrapInTransaction performs the given action wrapped in a transaction. Will Commit() on success and Rollback() on a non-nil error returned.
func (dbc *DbConnection) WrapInTransaction(action func(*sql.Tx) error) error {
	tx, err := dbc.Begin()
	if err != nil {
		return exception.Wrap(err)
	}
	err = action(tx)
	if err != nil {
		if rollbackErr := dbc.Rollback(tx); rollbackErr != nil {
			return exception.WrapMany(rollbackErr, err)
		}
		return exception.Wrap(err)
	} else if commitErr := dbc.Commit(tx); commitErr != nil {
		return exception.Wrap(commitErr)
	}
	return nil
}

// Prepare prepares a new statement for the connection.
func (dbc *DbConnection) Prepare(statement string, tx *sql.Tx) (*sql.Stmt, error) {
	if dbc == nil {
		return nil, exception.New(DBAliasNilError)
	}

	if tx != nil {
		stmt, stmtErr := tx.Prepare(statement)
		if stmtErr != nil {
			return nil, exception.Newf("Postgres Error: %v", stmtErr)
		}
		return stmt, nil
	}

	if dbc.Tx != nil {
		stmt, stmtErr := dbc.Tx.Prepare(statement)
		if stmtErr != nil {
			return nil, exception.Newf("Postgres Error: %v", stmtErr)
		}
		return stmt, nil
	}

	// open shared connection
	dbConn, dbErr := dbc.Open()
	if dbErr != nil {
		return nil, exception.Newf("Postgres Error: %v", dbErr)
	}
	stmt, stmtErr := dbConn.Prepare(statement)
	if stmtErr != nil {
		return nil, exception.Newf("Postgres Error: %v", stmtErr)
	}
	return stmt, nil
}

// OpenNew returns a new connection object.
func (dbc *DbConnection) OpenNew() (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", dbc.CreatePostgresConnectionString())
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return dbConn, nil
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbc *DbConnection) Open() (*sql.DB, error) {
	if dbc.Connection == nil {
		dbc.MetaLock.Lock()
		defer dbc.MetaLock.Unlock()

		if dbc.Connection == nil {
			newConn, err := dbc.OpenNew()
			if err != nil {
				return nil, exception.Wrap(err)
			}
			dbc.Connection = newConn
		}
	}
	return dbc.Connection, nil
}

// Exec runs the statement without creating a QueryResult.
func (dbc *DbConnection) Exec(statement string, args ...interface{}) error {
	return dbc.ExecInTransaction(statement, nil, args...)
}

// ExecInTransaction runs a statement within a transaction.
func (dbc *DbConnection) ExecInTransaction(statement string, tx *sql.Tx, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	stmt, stmtErr := dbc.Prepare(statement, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	if _, execErr := stmt.Exec(args...); execErr != nil {
		err = exception.Wrap(execErr)
		return
	}

	return
}

// Query runs the selected statement and returns a QueryResult.
func (dbc *DbConnection) Query(statement string, args ...interface{}) *QueryResult {
	return dbc.QueryInTransaction(statement, nil, args...)
}

// QueryInTransaction runs the selected statement in a transaction and returns a QueryResult.
func (dbc *DbConnection) QueryInTransaction(statement string, tx *sql.Tx, args ...interface{}) (result *QueryResult) {
	result = &QueryResult{conn: dbc}
	if dbc == nil {
		result.err = exception.New(DBAliasNilError)
		return
	}
	dbc.txLock()

	stmt, stmtErr := dbc.Prepare(statement, tx)
	if stmtErr != nil {
		result.err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			closeErr := stmt.Close()
			result.err = exception.WrapMany(result.err, exception.New(r), closeErr)
			dbc.txUnlock()
		}
	}()

	rows, queryErr := stmt.Query(args...)
	if queryErr != nil {
		result.err = exception.Wrap(queryErr)
		return
	}

	// the result MUST close these.
	result.stmt = stmt
	result.rows = rows
	return
}

// GetByID returns a given object based on a group of primary key ids.
func (dbc *DbConnection) GetByID(object DatabaseMapped, ids ...interface{}) error {
	return dbc.GetByIDInTransaction(object, nil, ids...)
}

// GetByIDInTransaction returns a given object based on a group of primary key ids within a transaction.
func (dbc *DbConnection) GetByIDInTransaction(object DatabaseMapped, tx *sql.Tx, ids ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := CachedColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	columnNames := standardCols.ColumnNames()
	tableName := object.TableName()
	pks := standardCols.PrimaryKeys()

	if pks.Len() == 0 {
		err = exception.New("no primary key on object to get by.")
		return
	}

	queryBody := fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columnNames, ","), tableName, makeWhereClause(pks, 1))

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	rows, queryErr := stmt.Query(ids...)
	if queryErr != nil {
		err = exception.Wrap(queryErr)
		return
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	var popErr error
	if rows.Next() {
		if IsPopulatable(object) {
			popErr = AsPopulatable(object).Populate(rows)
		} else {
			popErr = PopulateInOrder(object, rows, standardCols)
		}

		if popErr != nil {
			err = exception.Wrap(popErr)
			return
		}
	}

	err = exception.Wrap(rows.Err())
	return
}

// GetAll returns all rows of an object mapped table.
func (dbc *DbConnection) GetAll(collection interface{}) error {
	return dbc.GetAllInTransaction(collection, nil)
}

// GetAllInTransaction returns all rows of an object mapped table wrapped in a transaction.
func (dbc *DbConnection) GetAllInTransaction(collection interface{}, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)
	meta := CachedColumnCollectionFromType(tableName, t).NotReadOnly()

	columnNames := meta.ColumnNames()
	sqlStmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columnNames, ","), tableName)

	stmt, stmtErr := dbc.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	rows, queryErr := stmt.Query()
	if queryErr != nil {
		err = exception.Wrap(queryErr)
		return
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	v, _ := MakeNew(t)
	isPopulatable := IsPopulatable(v)

	var popErr error
	for rows.Next() {
		newObj, _ := MakeNew(t)

		if isPopulatable {
			popErr = AsPopulatable(newObj).Populate(rows)
		} else {
			popErr = PopulateInOrder(newObj, rows, meta)
			if popErr != nil {
				err = exception.Wrap(popErr)
				return
			}
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
	}

	err = exception.Wrap(rows.Err())
	return
}

// Create writes an object to the database.
func (dbc *DbConnection) Create(object DatabaseMapped) error {
	return dbc.CreateInTransaction(object, nil)
}

// CreateInTransaction writes an object to the database within a transaction.
func (dbc *DbConnection) CreateInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	cols := CachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()
	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)
	tokens := ParamTokensCSV(writeCols.Len())

	var sqlStmt string
	if serials.Len() == 0 {
		sqlStmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(colNames, ","),
			tokens,
		)
	} else {
		serial := serials.FirstOrDefault()
		sqlStmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			tableName,
			strings.Join(colNames, ","),
			tokens,
			serial.ColumnName,
		)
	}

	stmt, stmtErr := dbc.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	if serials.Len() == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			err = exception.Wrap(execErr)
			return
		}
	} else {
		serial := serials.FirstOrDefault()

		var id interface{}
		execErr := stmt.QueryRow(colValues...).Scan(&id)
		if execErr != nil {
			err = exception.Wrap(execErr)
			return
		}
		setErr := serial.SetValue(object, id)
		if setErr != nil {
			err = exception.Wrap(setErr)
			return
		}
	}

	return nil
}

// Update updates an object.
func (dbc *DbConnection) Update(object DatabaseMapped) error {
	return dbc.UpdateInTransaction(object, nil)
}

// UpdateInTransaction updates an object wrapped in a transaction.
func (dbc *DbConnection) UpdateInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials().NotPrimaryKeys()
	pks := cols.PrimaryKeys()
	allCols := writeCols.ConcatWith(pks)
	totalValues := allCols.ColumnValues(object)
	numColumns := writeCols.Len()

	sqlStmt := "UPDATE " + tableName + " SET "
	for i, col := range writeCols.Columns() {
		sqlStmt = sqlStmt + col.ColumnName + " = $" + strconv.Itoa(i+1)
		if i != numColumns-1 {
			sqlStmt = sqlStmt + ","
		}
	}

	whereClause := makeWhereClause(pks, numColumns+1)
	sqlStmt = sqlStmt + whereClause

	stmt, stmtErr := dbc.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	_, execErr := stmt.Exec(totalValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbc *DbConnection) Exists(object DatabaseMapped) (bool, error) {
	return dbc.ExistsInTransaction(object, nil)
}

// ExistsInTransaction returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbc *DbConnection) ExistsInTransaction(object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return false, exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if pks.Len() == 0 {
		exists = false
		err = exception.New("No primary key on object.")
		return
	}
	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("SELECT 1 FROM %s %s", tableName, whereClause)
	stmt, stmtErr := dbc.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		exists = false
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	pkValues := pks.ColumnValues(object)
	rows, queryErr := stmt.Query(pkValues...)
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	if queryErr != nil {
		exists = false
		err = exception.Wrap(queryErr)
		return
	}

	exists = rows.Next()
	return
}

// Delete deletes an object from the database.
func (dbc *DbConnection) Delete(object DatabaseMapped) error {
	return dbc.DeleteInTransaction(object, nil)
}

// DeleteInTransaction deletes an object from the database wrapped in a transaction.
func (dbc *DbConnection) DeleteInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.txLock()
	defer dbc.txUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("DELETE FROM %s %s", tableName, whereClause)

	stmt, stmtErr := dbc.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		closeErr := stmt.Close()
		if closeErr != nil {
			err = exception.WrapMany(err, closeErr)
		}
	}()

	pkValues := pks.ColumnValues(object)

	_, execErr := stmt.Exec(pkValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
	}
	return
}

// IsolateToTransaction causes all commands on the given connection to use a transaction.
// NOTE: causes locking around the transaction.
func (dbc *DbConnection) IsolateToTransaction(tx *sql.Tx) {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	dbc.TxLock.Lock()
	defer dbc.TxLock.Unlock()
	dbc.Tx = tx
}

// ReleaseIsolation reverses `IsolateToTransaction`
func (dbc *DbConnection) ReleaseIsolation() {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	dbc.TxLock.Lock()
	defer dbc.TxLock.Unlock()
	dbc.Tx = nil
}

// IsIsolatedToTransaction returns if the connection is isolated to a transaction.
func (dbc *DbConnection) IsIsolatedToTransaction() bool {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	dbc.TxLock.RLock()
	defer dbc.TxLock.RUnlock()

	return dbc.Tx != nil
}

// Commit commits a transaction if the connection is not currently isolated to one already.
func (dbc *DbConnection) Commit(tx *sql.Tx) error {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return nil
	}
	return tx.Commit()
}

// Rollback commits a transaction if the connection is not currently isolated to one already.
func (dbc *DbConnection) Rollback(tx *sql.Tx) error {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return nil
	}
	return tx.Rollback()
}

func (dbc *DbConnection) txLock() {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	if dbc.Tx != nil {
		dbc.TxLock.Lock()
	}
}

func (dbc *DbConnection) txUnlock() {
	if dbc == nil {
		panic(DBAliasNilError)
	}

	if dbc.Tx != nil {
		dbc.TxLock.Unlock()
	}
}
