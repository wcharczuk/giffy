// Package spiffy providers a basic abstraction layer above normal database/sql that makes it easier to
// interact with the database and organize database related code. It is not intended to replace actual sql
// (you write queries yourself).
package spiffy

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"

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

// NewDbConnection returns a new DbConnectin.
func NewDbConnection() *DbConnection {
	return &DbConnection{
		useStatementCache:      false, //doesnt actually help perf, maybe someday.
		statementCacheInitLock: &sync.Mutex{},
		connectionLock:         &sync.Mutex{},
		txLock:                 &sync.RWMutex{},
	}
}

// NewDbConnectionWithHost creates a new DbConnection using current user peer authentication.
func NewDbConnectionWithHost(host, dbName string) *DbConnection {
	dbc := NewDbConnection()
	dbc.Host = host
	dbc.Database = dbName
	dbc.SSLMode = "disable"
	return dbc
}

// NewDbConnectionWithPassword creates a new connection with SSLMode set to "disable"
func NewDbConnectionWithPassword(host, dbName, username, password string) *DbConnection {
	dbc := NewDbConnection()
	dbc.Host = host
	dbc.Database = dbName
	dbc.Username = username
	dbc.Password = password
	dbc.SSLMode = "disable"
	return dbc
}

// NewDbConnectionWithSSLMode creates a new connection with all available options (including SSLMode)
func NewDbConnectionWithSSLMode(host, dbName, username, password, sslMode string) *DbConnection {
	dbc := NewDbConnection()
	dbc.Host = host
	dbc.Database = dbName
	dbc.Username = username
	dbc.Password = password
	dbc.SSLMode = sslMode
	return dbc
}

// NewDbConnectionFromDSN creates a new connection with SSLMode set to "disable"
func NewDbConnectionFromDSN(dsn string) *DbConnection {
	dbc := NewDbConnection()
	dbc.DSN = dsn
	return dbc
}

func envVarWithDefault(varName, defaultValue string) string {
	envVarValue := os.Getenv(varName)
	if len(envVarValue) > 0 {
		return envVarValue
	}
	return defaultValue
}

// NewDbConnectionFromEnvironment creates a new db connection from environment variables.
//
// The environment variable mappings are as follows:
//	-	DATABSE_URL 	= DSN 	//note that this trumps other vars (!!)
// 	-	DB_HOST 		= Host
//	-	DB_PORT 		= Port
//	- 	DB_NAME 		= Database
//	-	DB_SCHEMA		= Schema
//	-	DB_USER 		= Username
//	-	DB_PASSWORD 	= Password
//	-	DB_SSLMODE 		= SSLMode
func NewDbConnectionFromEnvironment() *DbConnection {
	if len(os.Getenv("DATABASE_URL")) > 0 {
		return NewDbConnectionFromDSN(os.Getenv("DATABASE_URL"))
	}

	dbc := NewDbConnection()
	dbc.Host = envVarWithDefault("DB_HOST", "localhost")
	dbc.Database = os.Getenv("DB_NAME")
	dbc.Schema = os.Getenv("DB_SCHEMA")
	dbc.Username = os.Getenv("DB_USER")
	dbc.Password = os.Getenv("DB_PASSWORD")
	dbc.SSLMode = envVarWithDefault("DB_SSLMODE", "disable")
	return dbc
}

// DbConnection is the basic wrapper for connection parameters and saves a reference to the created sql.Connection.
type DbConnection struct {
	// DSN is a fully formed DSN (this skips DSN formation from other variables).
	DSN string

	// Host is the server to connect to.
	Host string
	// Port is the port to connect to.
	Port string
	// DBName is the database name
	Database string
	// Schema is the application schema within the database, defaults to `public`.
	Schema string
	// Username is the username for the connection via password auth.
	Username string
	// Password is the password for the connection via password auth.
	Password string
	// SSLMode is the sslmode for the connection.
	SSLMode string

	Connection *sql.DB

	tx *sql.Tx

	connectionLock         *sync.Mutex
	txLock                 *sync.RWMutex
	statementCacheInitLock *sync.Mutex

	diagnostics *logger.DiagnosticsAgent

	useStatementCache bool
	statementCache    *StatementCache
}

// Close implements a closer.
func (dbc *DbConnection) Close() error {
	var err error
	if dbc.statementCache != nil {
		err = dbc.statementCache.Close()
	}
	if err != nil {
		return err
	}
	return dbc.Connection.Close()
}

// SetDiagnostics sets the connection's diagnostic agent.
func (dbc *DbConnection) SetDiagnostics(agent *logger.DiagnosticsAgent) {
	dbc.diagnostics = agent
}

// Diagnostics returns the diagnostics agent.
func (dbc *DbConnection) Diagnostics() *logger.DiagnosticsAgent {
	return dbc.diagnostics
}

func (dbc *DbConnection) fireEvent(flag logger.EventFlag, query string, elapsed time.Duration, err error) {
	if dbc.diagnostics != nil {
		dbc.diagnostics.OnEvent(flag, query, elapsed, err)
	}
}

// EnableStatementCache opts to cache statements for the connection.
func (dbc *DbConnection) EnableStatementCache() {
	dbc.useStatementCache = true
}

// DisableStatementCache opts to not use the statement cache.
func (dbc *DbConnection) DisableStatementCache() {
	dbc.useStatementCache = false
}

// StatementCache returns the statement cache.
func (dbc *DbConnection) StatementCache() *StatementCache {
	return dbc.statementCache
}

// CreatePostgresConnectionString returns a sql connection string from a given set of DbConnection parameters.
func (dbc *DbConnection) CreatePostgresConnectionString() (string, error) {
	if len(dbc.DSN) != 0 {
		return dbc.DSN, nil
	}

	if len(dbc.Database) == 0 {
		return "", exception.New("`DB_NAME` is required to open a new connection")
	}

	sslMode := "?sslmode=disable"
	if len(dbc.SSLMode) > 0 {
		sslMode = fmt.Sprintf("?sslmode=%s", url.QueryEscape(dbc.SSLMode))
	}

	var portSegment string
	if len(dbc.Port) > 0 {
		portSegment = fmt.Sprintf(":%s", dbc.Port)
	}

	if dbc.Username != "" {
		if dbc.Password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s%s/%s%s", url.QueryEscape(dbc.Username), url.QueryEscape(dbc.Password), dbc.Host, portSegment, dbc.Database, sslMode), nil
		}
		return fmt.Sprintf("postgres://%s@%s%s/%s%s", url.QueryEscape(dbc.Username), dbc.Host, portSegment, dbc.Database, sslMode), nil
	}
	return fmt.Sprintf("postgres://%s%s/%s%s", dbc.Host, portSegment, dbc.Database, sslMode), nil
}

// Begin starts a new transaction.
func (dbc *DbConnection) Begin() (*sql.Tx, error) {
	if dbc == nil {
		return nil, exception.New(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return dbc.tx, nil
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

// WrapInTx performs the given action wrapped in a transaction. Will Commit() on success and Rollback() on a non-nil error returned.
func (dbc *DbConnection) WrapInTx(action func(*sql.Tx) error) error {
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
		stmt, err := tx.Prepare(statement)
		if err != nil {
			return nil, exception.Newf("Postgres Error: %v", err)
		}
		return stmt, nil
	}

	if dbc.tx != nil {
		stmt, err := dbc.tx.Prepare(statement)
		if err != nil {
			return nil, exception.Newf("Postgres Error: %v", err)
		}
		return stmt, nil
	}

	// open shared connection
	dbConn, err := dbc.Open()
	if err != nil {
		return nil, exception.Newf("Postgres Error: %v", err)
	}

	if dbc.useStatementCache {
		if dbc.statementCache == nil {
			dbc.statementCacheInitLock.Lock()
			defer dbc.statementCacheInitLock.Unlock()
			if dbc.statementCache == nil {
				dbc.statementCache = NewStatementCache(dbConn)
			}
		}
		return dbc.statementCache.Prepare(statement)
	}

	stmt, err := dbConn.Prepare(statement)
	if err != nil {
		return nil, exception.Newf("Postgres Error: %v", err)
	}
	return stmt, nil
}

// OpenNew returns a new connection object.
func (dbc *DbConnection) OpenNew() (*sql.DB, error) {

	connStr, err := dbc.CreatePostgresConnectionString()
	if err != nil {
		return nil, err
	}

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	if len(dbc.Schema) > 0 {
		_, err = dbConn.Exec(fmt.Sprintf("SET search_path TO %s,public;", dbc.Schema))
		if err != nil {
			return nil, err
		}
	}

	return dbConn, nil
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbc *DbConnection) Open() (*sql.DB, error) {
	if dbc.Connection == nil {
		dbc.connectionLock.Lock()
		defer dbc.connectionLock.Unlock()

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
	return dbc.ExecInTx(statement, nil, args...)
}

// ExecInTx runs a statement within a transaction.
func (dbc *DbConnection) ExecInTx(statement string, tx *sql.Tx, args ...interface{}) (err error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagExecute, statement, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	stmt, stmtErr := dbc.Prepare(statement, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			closeErr := stmt.Close()
			if closeErr != nil {
				err = exception.WrapMany(err, closeErr)
			}
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
	return dbc.QueryInTx(statement, nil, args...)
}

// QueryInTx runs the selected statement in a transaction and returns a QueryResult.
func (dbc *DbConnection) QueryInTx(statement string, tx *sql.Tx, args ...interface{}) (result *QueryResult) {
	result = &QueryResult{queryBody: statement, start: time.Now(), conn: dbc}
	if dbc == nil {
		result.err = exception.New(DBAliasNilError)
		return
	}
	dbc.transactionLock()

	stmt, stmtErr := dbc.Prepare(statement, tx)
	if stmtErr != nil {
		result.err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if !dbc.useStatementCache {
				result.err = exception.WrapMany(result.err, exception.New(r), stmt.Close())
			} else {
				result.err = exception.WrapMany(result.err, exception.New(r))
			}

			dbc.transactionUnlock()
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
	return dbc.GetByIDInTx(object, nil, ids...)
}

// GetByIDInTx returns a given object based on a group of primary key ids within a transaction.
func (dbc *DbConnection) GetByIDInTx(object DatabaseMapped, tx *sql.Tx, ids ...interface{}) (err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagExecute, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

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

	queryBody = fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columnNames, ","), tableName, makeWhereClause(pks, 1))

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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
	return dbc.GetAllInTx(collection, nil)
}

// GetAllInTx returns all rows of an object mapped table wrapped in a transaction.
func (dbc *DbConnection) GetAllInTx(collection interface{}, tx *sql.Tx) (err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagQuery, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)
	meta := CachedColumnCollectionFromType(tableName, t).NotReadOnly()

	columnNames := meta.ColumnNames()
	queryBody = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columnNames, ","), tableName)

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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

	v, err := MakeNewDatabaseMapped(t)
	if err != nil {
		return
	}
	isPopulatable := IsPopulatable(v)

	var popErr error
	for rows.Next() {
		newObj, _ := MakeNewDatabaseMapped(t)

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
	return dbc.CreateInTx(object, nil)
}

// CreateInTx writes an object to the database within a transaction.
func (dbc *DbConnection) CreateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagExecute, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	cols := CachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()
	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)
	tokens := ParamTokensCSV(writeCols.Len())

	if serials.Len() == 0 {
		queryBody = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(colNames, ","),
			tokens,
		)
	} else {
		serial := serials.FirstOrDefault()
		queryBody = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			tableName,
			strings.Join(colNames, ","),
			tokens,
			serial.ColumnName,
		)
	}

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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
	return dbc.UpdateInTx(object, nil)
}

// UpdateInTx updates an object wrapped in a transaction.
func (dbc *DbConnection) UpdateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagExecute, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials().NotPrimaryKeys()
	pks := cols.PrimaryKeys()
	allCols := writeCols.ConcatWith(pks)
	totalValues := allCols.ColumnValues(object)
	numColumns := writeCols.Len()

	queryBody = "UPDATE " + tableName + " SET "
	for i, col := range writeCols.Columns() {
		queryBody = queryBody + col.ColumnName + " = $" + strconv.Itoa(i+1)
		if i != numColumns-1 {
			queryBody = queryBody + ","
		}
	}

	whereClause := makeWhereClause(pks, numColumns+1)
	queryBody = queryBody + whereClause

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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
	return dbc.ExistsInTx(object, nil)
}

// ExistsInTx returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbc *DbConnection) ExistsInTx(object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagQuery, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return false, exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if pks.Len() == 0 {
		exists = false
		err = exception.New("No primary key on object.")
		return
	}
	whereClause := makeWhereClause(pks, 1)
	queryBody = fmt.Sprintf("SELECT 1 FROM %s %s", tableName, whereClause)
	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		exists = false
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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
	return dbc.DeleteInTx(object, nil)
}

// DeleteInTx deletes an object from the database wrapped in a transaction.
func (dbc *DbConnection) DeleteInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	var queryBody string
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
		dbc.fireEvent(EventFlagExecute, queryBody, time.Now().Sub(start), err)
	}()

	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	dbc.transactionLock()
	defer dbc.transactionUnlock()

	tableName := object.TableName()
	cols := CachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	whereClause := makeWhereClause(pks, 1)
	queryBody = fmt.Sprintf("DELETE FROM %s %s", tableName, whereClause)

	stmt, stmtErr := dbc.Prepare(queryBody, tx)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if !dbc.useStatementCache {
			err = exception.WrapMany(err, stmt.Close())
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
		return
	}

	dbc.txLock.Lock()
	defer dbc.txLock.Unlock()
	dbc.tx = tx
}

// ReleaseIsolation reverses `IsolateToTransaction`
func (dbc *DbConnection) ReleaseIsolation() {
	if dbc == nil {
		return
	}

	dbc.txLock.Lock()
	defer dbc.txLock.Unlock()
	dbc.tx = nil
}

// IsIsolatedToTransaction returns if the connection is isolated to a transaction.
func (dbc *DbConnection) IsIsolatedToTransaction() bool {
	if dbc == nil {
		return false
	}

	dbc.txLock.RLock()
	defer dbc.txLock.RUnlock()

	return dbc.tx != nil
}

// Commit commits a transaction if the connection is not currently isolated to one already.
func (dbc *DbConnection) Commit(tx *sql.Tx) error {
	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return nil
	}
	return tx.Commit()
}

// Rollback commits a transaction if the connection is not currently isolated to one already.
func (dbc *DbConnection) Rollback(tx *sql.Tx) error {
	if dbc == nil {
		return exception.New(DBAliasNilError)
	}

	if dbc.IsIsolatedToTransaction() {
		return nil
	}
	return tx.Rollback()
}

func (dbc *DbConnection) transactionLock() {
	if dbc == nil {
		return
	}

	if dbc.tx != nil {
		dbc.txLock.Lock()
	}
}

func (dbc *DbConnection) transactionUnlock() {
	if dbc == nil {
		return
	}

	if dbc.tx != nil {
		dbc.txLock.Unlock()
	}
}
