// Package db providers a basic abstraction layer above normal database/sql that makes it easier to
// interact with the database and organize database related code. It is not intended to replace actual sql
// (you write queries yourself).
package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"

	// PQ is the postgres driver
	_ "github.com/lib/pq"
)

const (
	//DBNilError is a common error
	DBNilError = "connection is nil"
)

const (
	runeComma   = rune(',')
	runeNewline = rune('\n')
	runeTab     = rune('\t')
	runeSpace   = rune(' ')
)

// --------------------------------------------------------------------------------
// Connection
// --------------------------------------------------------------------------------

// New returns a new Connection.
func New() *Connection {
	return &Connection{
		Config:             &Config{},
		bufferPool:         NewBufferPool(DefaultBufferPoolSize),
		useStatementCache:  DefaultUseStatementCache,
		statementCacheLock: &sync.Mutex{},
		connectionLock:     &sync.Mutex{},
	}
}

// NewFromConfig returns a new connection from a config.
func NewFromConfig(cfg *Config) *Connection {
	return &Connection{
		Config:             cfg,
		bufferPool:         NewBufferPool(cfg.GetBufferPoolSize()),
		useStatementCache:  cfg.GetUseStatementCache(), //doesnt actually help perf, maybe someday.
		statementCacheLock: &sync.Mutex{},
		connectionLock:     &sync.Mutex{},
	}
}

// NewFromEnv creates a new db connection from environment variables.
func NewFromEnv() *Connection {
	return NewFromConfig(NewConfigFromEnv())
}

// Connection is the basic wrapper for connection parameters and saves a reference to the created sql.Connection.
type Connection struct {
	// Connection is the underlying sql driver connection for the Connection.
	Connection *sql.DB
	Config     *Config

	connectionLock     *sync.Mutex
	statementCacheLock *sync.Mutex

	bufferPool *BufferPool
	log        *logger.Logger

	useStatementCache bool
	statementCache    *StatementCache

	// database is used for logging.
	database string
}

// Database returns the connected database name.
func (dbc *Connection) Database() string {
	return dbc.database
}

// Close implements a closer.
func (dbc *Connection) Close() error {
	var err error
	if dbc.statementCache != nil {
		err = dbc.statementCache.Close()
	}
	if err != nil {
		return err
	}
	return dbc.Connection.Close()
}

// WithLogger sets the connection's diagnostic agent.
func (dbc *Connection) WithLogger(log *logger.Logger) {
	dbc.log = log
}

// Logger returns the diagnostics agent.
func (dbc *Connection) Logger() *logger.Logger {
	return dbc.log
}

func (dbc *Connection) fireEvent(flag logger.Flag, query string, elapsed time.Duration, err error, optionalQueryLabel ...string) {
	if dbc.log != nil {
		var queryLabel string
		if len(optionalQueryLabel) > 0 {
			queryLabel = optionalQueryLabel[0]
		}

		dbc.log.Trigger(logger.NewQueryEvent(query, elapsed).WithFlag(flag).WithDatabase(dbc.database).WithQueryLabel(queryLabel).WithEngine("postgres"))
		if err != nil {
			dbc.log.Error(err)
		}
	}
}

// EnableStatementCache opts to cache statements for the connection.
func (dbc *Connection) EnableStatementCache() {
	dbc.useStatementCache = true
}

// DisableStatementCache opts to not use the statement cache.
func (dbc *Connection) DisableStatementCache() {
	dbc.useStatementCache = false
}

// WithUseStatementCache returns if we should use the statement cache.
func (dbc *Connection) WithUseStatementCache(enabled bool) *Connection {
	dbc.useStatementCache = enabled
	return dbc
}

// StatementCache returns the statement cache.
func (dbc *Connection) StatementCache() *StatementCache {
	return dbc.statementCache
}

// openNewSQLConnection returns a new connection object.
func (dbc *Connection) openNewSQLConnection() (*sql.DB, error) {
	if dbc.Config == nil {
		return nil, exception.New("connection configuration is unset")
	}

	// the config resolution step is a little weird
	// first, fully synthesize the dsn
	// as it can be set directly or composed from individual fields
	dsn := dbc.Config.CreateDSN()
	// then re-parse it to get relevant fields we might want to save
	// like the database name etc.
	parsed, err := NewConfigFromDSN(dsn)
	if err != nil {
		exception.Wrap(err)
	}

	// open the connection
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	// memoize the db name for logging calls
	dbc.database = parsed.GetDatabase()
	// action config points.
	dbConn.SetConnMaxLifetime(dbc.Config.GetMaxLifetime())
	dbConn.SetMaxIdleConns(dbc.Config.GetIdleConnections())
	dbConn.SetMaxOpenConns(dbc.Config.GetMaxConnections())

	schema := dbc.Config.GetSchema()
	if len(schema) > 0 {
		_, err = dbConn.Exec(fmt.Sprintf("SET search_path TO %s,public;", schema))
		if err != nil {
			return nil, exception.Wrap(err)
		}
	}

	// sanity check on the connection.
	_, err = dbConn.Exec("select 'ok!'")
	if err != nil {
		return nil, exception.Wrap(err)
	}

	return dbConn, nil
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbc *Connection) Open() (*Connection, error) {
	if dbc.Connection == nil {
		dbc.connectionLock.Lock()
		defer dbc.connectionLock.Unlock()

		if dbc.Connection == nil {
			newConn, err := dbc.openNewSQLConnection()
			if err != nil {
				return nil, err
			}
			dbc.Connection = newConn
		}
	}
	return dbc, nil
}

// Begin starts a new transaction.
func (dbc *Connection) Begin() (*sql.Tx, error) {
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

// Prepare prepares a new statement for the connection.
func (dbc *Connection) Prepare(statement string, tx *sql.Tx) (*sql.Stmt, error) {
	if tx != nil {
		stmt, err := tx.Prepare(statement)
		if err != nil {
			return nil, exception.Wrap(err)
		}
		return stmt, nil
	}

	// open shared connection
	dbConn, err := dbc.Open()
	if err != nil {
		return nil, exception.Wrap(err)
	}

	stmt, err := dbConn.Connection.Prepare(statement)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return stmt, nil
}

func (dbc *Connection) ensureStatementCache() error {
	if dbc.statementCache == nil {
		dbc.statementCacheLock.Lock()
		defer dbc.statementCacheLock.Unlock()
		if dbc.statementCache == nil {
			db, err := dbc.Open()
			if err != nil {
				return exception.Wrap(err)
			}
			dbc.statementCache = newStatementCache(db.Connection)
		}
	}
	return nil
}

// PrepareCached prepares a potentially cached statement.
func (dbc *Connection) PrepareCached(id, statement string, tx *sql.Tx) (*sql.Stmt, error) {
	if tx != nil {
		stmt, err := tx.Prepare(statement)
		if err != nil {
			return nil, exception.Wrap(err)
		}
		return stmt, nil
	}

	if dbc.useStatementCache {
		dbc.ensureStatementCache()
		return dbc.statementCache.Prepare(id, statement)
	}
	return dbc.Prepare(statement, tx)
}

// --------------------------------------------------------------------------------
// Invocation context
// --------------------------------------------------------------------------------

// DB returns a new db context.
func (dbc *Connection) DB(txs ...*sql.Tx) *DB {
	return &DB{
		conn:       dbc,
		tx:         OptionalTx(txs...),
		fireEvents: dbc.log != nil,
	}
}

// Invoke returns a new invocation.
func (dbc *Connection) Invoke(txs ...*sql.Tx) *Invocation {
	return &Invocation{
		conn:       dbc,
		tx:         OptionalTx(txs...),
		fireEvents: dbc.log != nil,
	}
}

// InTx is an alias to Invoke.
func (dbc *Connection) InTx(txs ...*sql.Tx) *Invocation {
	return dbc.Invoke(txs...)
}

// --------------------------------------------------------------------------------
// Invocation Context Stubs
// --------------------------------------------------------------------------------

// Exec runs the statement without creating a QueryResult.
func (dbc *Connection) Exec(statement string, args ...interface{}) error {
	return dbc.ExecInTx(statement, nil, args...)
}

// ExecWithCacheLabel runs the statement without creating a QueryResult.
func (dbc *Connection) ExecWithCacheLabel(statement, cacheLabel string, args ...interface{}) error {
	return dbc.ExecInTxWithCacheLabel(statement, cacheLabel, nil, args...)
}

// ExecInTx runs a statement within a transaction.
func (dbc *Connection) ExecInTx(statement string, tx *sql.Tx, args ...interface{}) (err error) {
	return dbc.ExecInTxWithCacheLabel(statement, statement, tx, args...)
}

// ExecInTxWithCacheLabel runs a statement within a transaction.
func (dbc *Connection) ExecInTxWithCacheLabel(statement, cacheLabel string, tx *sql.Tx, args ...interface{}) (err error) {
	return dbc.Invoke(tx).WithLabel(cacheLabel).Exec(statement, args...)
}

// Query runs the selected statement and returns a Query.
func (dbc *Connection) Query(statement string, args ...interface{}) *Query {
	return dbc.QueryInTx(statement, nil, args...)
}

// QueryInTx runs the selected statement in a transaction and returns a Query.
func (dbc *Connection) QueryInTx(statement string, tx *sql.Tx, args ...interface{}) (result *Query) {
	return dbc.Invoke(tx).Query(statement, args...)
}

// Get returns a given object based on a group of primary key ids.
func (dbc *Connection) Get(object DatabaseMapped, ids ...interface{}) error {
	return dbc.GetInTx(object, nil, ids...)
}

// GetInTx returns a given object based on a group of primary key ids within a transaction.
func (dbc *Connection) GetInTx(object DatabaseMapped, tx *sql.Tx, args ...interface{}) error {
	return dbc.Invoke(tx).Get(object, args...)
}

// GetAll returns all rows of an object mapped table.
func (dbc *Connection) GetAll(collection interface{}) error {
	return dbc.GetAllInTx(collection, nil)
}

// GetAllInTx returns all rows of an object mapped table wrapped in a transaction.
func (dbc *Connection) GetAllInTx(collection interface{}, tx *sql.Tx) error {
	return dbc.Invoke(tx).GetAll(collection)
}

// Create writes an object to the database.
func (dbc *Connection) Create(object DatabaseMapped) error {
	return dbc.CreateInTx(object, nil)
}

// CreateInTx writes an object to the database within a transaction.
func (dbc *Connection) CreateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).Create(object)
}

// CreateIfNotExists writes an object to the database if it does not already exist.
func (dbc *Connection) CreateIfNotExists(object DatabaseMapped) error {
	return dbc.CreateIfNotExistsInTx(object, nil)
}

// CreateIfNotExistsInTx writes an object to the database if it does not already exist within a transaction.
func (dbc *Connection) CreateIfNotExistsInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).CreateIfNotExists(object)
}

// CreateMany writes many an objects to the database.
func (dbc *Connection) CreateMany(objects interface{}) error {
	return dbc.CreateManyInTx(objects, nil)
}

// CreateManyInTx writes many an objects to the database within a transaction.
func (dbc *Connection) CreateManyInTx(objects interface{}, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).CreateMany(objects)
}

// Update updates an object.
func (dbc *Connection) Update(object DatabaseMapped) error {
	return dbc.UpdateInTx(object, nil)
}

// UpdateInTx updates an object wrapped in a transaction.
func (dbc *Connection) UpdateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).Update(object)
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbc *Connection) Exists(object DatabaseMapped) (bool, error) {
	return dbc.ExistsInTx(object, nil)
}

// ExistsInTx returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbc *Connection) ExistsInTx(object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	return dbc.Invoke(tx).Exists(object)
}

// Delete deletes an object from the database.
func (dbc *Connection) Delete(object DatabaseMapped) error {
	return dbc.DeleteInTx(object, nil)
}

// DeleteInTx deletes an object from the database wrapped in a transaction.
func (dbc *Connection) DeleteInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).Delete(object)
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it.
func (dbc *Connection) Upsert(object DatabaseMapped) error {
	return dbc.UpsertInTx(object, nil)
}

// UpsertInTx inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (dbc *Connection) UpsertInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(tx).Upsert(object)
}

// Truncate fully removes an tables rows in a single opertation.
func (dbc *Connection) Truncate(object DatabaseMapped) error {
	return dbc.TruncateInTx(object, nil)
}

// TruncateInTx applies a truncation in a transaction.
func (dbc *Connection) TruncateInTx(object DatabaseMapped, tx *sql.Tx) error {
	return dbc.Invoke(tx).Truncate(object)
}
