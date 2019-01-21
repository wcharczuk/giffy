package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/lib/pq"

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
// It will use very bare bones defaults for the config.
func New() *Connection {
	return &Connection{
		config:    &Config{},
		planCache: NewPlanCache(),
	}
}

// NewFromConfig returns a new connection from a config.
func NewFromConfig(cfg *Config) (*Connection, error) {
	dsn := cfg.CreateDSN()
	parsed, err := NewConfigFromDSN(dsn)
	if err != nil {
		return nil, err
	}
	return New().WithConfig(parsed), nil
}

// MustNewFromConfig returns a new connection from a config
// and panics if there is an error.
func MustNewFromConfig(cfg *Config) *Connection {
	conn, err := NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
	return conn
}

// MustNewFromEnv creates a new db connection from environment variables.
// It will panic if there is an error.
func MustNewFromEnv() *Connection {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		panic(err)
	}
	return MustNewFromConfig(cfg)
}

// NewFromEnv will returns a new connection from a config
// set from environment variables.
func NewFromEnv() (*Connection, error) {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewFromConfig(cfg)
}

// Connection is the basic wrapper for connection parameters and saves a reference to the created sql.Connection.
type Connection struct {
	sync.Mutex
	tracer               Tracer
	statementInterceptor StatementInterceptor

	connection *sql.DB
	config     *Config
	bufferPool *BufferPool
	log        logger.Log
	planCache  *PlanCache
}

// WithConfig sets the config.
func (dbc *Connection) WithConfig(cfg *Config) *Connection {
	dbc.config = cfg
	return dbc
}

// Config returns the config.
func (dbc *Connection) Config() *Config {
	return dbc.config
}

// WithTracer sets the connection tracer and returns a reference.
func (dbc *Connection) WithTracer(tracer Tracer) *Connection {
	dbc.tracer = tracer
	return dbc
}

// Tracer returns the tracer.
func (dbc *Connection) Tracer() Tracer {
	return dbc.tracer
}

// WithStatementInterceptor sets the connection statement interceptor.
func (dbc *Connection) WithStatementInterceptor(interceptor StatementInterceptor) *Connection {
	dbc.statementInterceptor = interceptor
	return dbc
}

// StatementInterceptor returns the statement interceptor.
func (dbc *Connection) StatementInterceptor() StatementInterceptor {
	return dbc.statementInterceptor
}

// Connection returns the underlying driver connection.
func (dbc *Connection) Connection() *sql.DB {
	return dbc.connection
}

// Close implements a closer.
func (dbc *Connection) Close() error {
	if dbc.planCache != nil {
		if err := dbc.planCache.Close(); err != nil {
			return err
		}
	}
	return dbc.connection.Close()
}

// WithLogger sets the connection's diagnostic agent.
func (dbc *Connection) WithLogger(log logger.Log) *Connection {
	dbc.log = log
	return dbc
}

// Logger returns the diagnostics agent.
func (dbc *Connection) Logger() logger.Log {
	return dbc.log
}

// PlanCache returns the statement cache.
func (dbc *Connection) PlanCache() *PlanCache {
	return dbc.planCache
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbc *Connection) Open() error {
	dbc.Lock()
	defer dbc.Unlock()

	// bail if we've already opened the connection.
	if dbc.connection != nil {
		return Error(ErrConnectionAlreadyOpen)
	}
	if dbc.config == nil {
		return Error(ErrConfigUnset)
	}
	if dbc.bufferPool == nil {
		dbc.bufferPool = NewBufferPool(dbc.config.GetBufferPoolSize())
	}
	if dbc.planCache == nil {
		dbc.planCache = NewPlanCache()
	}

	dsn := dbc.config.CreateDSN()
	namedValues, err := pq.ParseURL(dsn)
	if err != nil {
		return err
	}

	// open the connection
	dbConn, err := sql.Open(dbc.config.GetEngine(), namedValues)
	if err != nil {
		return Error(err)
	}

	dbc.planCache.WithConnection(dbConn)
	dbc.planCache.WithEnabled(!dbc.config.GetPlanCacheDisabled())

	dbc.connection = dbConn
	dbc.connection.SetConnMaxLifetime(dbc.config.GetMaxLifetime())
	dbc.connection.SetMaxIdleConns(dbc.config.GetIdleConnections())
	dbc.connection.SetMaxOpenConns(dbc.config.GetMaxConnections())
	return nil
}

// Begin starts a new transaction.
func (dbc *Connection) Begin(opts ...*sql.TxOptions) (*sql.Tx, error) {
	return dbc.BeginContext(context.Background(), opts...)
}

// BeginContext starts a new transaction in a givent context.
func (dbc *Connection) BeginContext(context context.Context, opts ...*sql.TxOptions) (*sql.Tx, error) {
	if len(opts) > 0 {
		tx, err := dbc.connection.BeginTx(context, opts[0])
		return tx, Error(err)
	}
	tx, err := dbc.connection.BeginTx(context, nil)
	return tx, Error(err)
}

// PrepareContext prepares a statement potentially returning a cached version of the statement.
func (dbc *Connection) PrepareContext(context context.Context, cachedPlanKey, statement string, txs ...*sql.Tx) (stmt *sql.Stmt, err error) {
	if dbc.tracer != nil {
		tf := dbc.tracer.Prepare(context, dbc, statement)
		if tf != nil {
			defer func() { tf.Finish(err) }()
		}
	}
	if tx := Tx(txs...); tx != nil {
		stmt, err = tx.PrepareContext(context, statement)
		return
	}
	if dbc.planCache != nil && dbc.planCache.Enabled() && cachedPlanKey != "" {
		stmt, err = dbc.planCache.PrepareContext(context, cachedPlanKey, statement)
		return
	}
	stmt, err = dbc.connection.PrepareContext(context, statement)
	return
}

// --------------------------------------------------------------------------------
// Invocation
// --------------------------------------------------------------------------------

// Invoke returns a new invocation.
func (dbc *Connection) Invoke(context context.Context, txs ...*sql.Tx) *Invocation {
	return &Invocation{
		context:              context,
		tracer:               dbc.tracer,
		statementInterceptor: dbc.statementInterceptor,
		conn:                 dbc,
		startTime:            time.Now().UTC(),
		tx:                   OptionalTx(txs...),
	}
}

// Background returns an empty context.Context.
func (dbc *Connection) Background() context.Context {
	return context.Background()
}

// Ping checks the db connection.
func (dbc *Connection) Ping() error {
	return Error(dbc.connection.Ping())
}

// PingContext checks the db connection.
func (dbc *Connection) PingContext(context context.Context) (err error) {
	if dbc.tracer != nil {
		tf := dbc.tracer.Ping(context, dbc)
		if tf != nil {
			defer func() {
				tf.Finish(err)
			}()
		}
	}

	err = Error(dbc.connection.PingContext(context))
	return
}

// --------------------------------------------------------------------------------
// Invocation Context Stubs
//
// These are stubs that both preserve backwards compatibility but also help
// incrementally add extra functionality without needing to dig into the
// invocation fluent api.
// --------------------------------------------------------------------------------

// Exec runs the statement without creating a QueryResult.
func (dbc *Connection) Exec(statement string, args ...interface{}) error {
	return dbc.Invoke(dbc.Background()).Exec(statement, args...)
}

// ExecContext runs the statement without creating a QueryResult.
func (dbc *Connection) ExecContext(context context.Context, statement string, args ...interface{}) error {
	return dbc.Invoke(context).Exec(statement, args...)
}

// ExecWithCachedPlan runs the statement without creating a QueryResult.
func (dbc *Connection) ExecWithCachedPlan(statement, cachedPlanKey string, args ...interface{}) error {
	return dbc.Invoke(context.Background()).WithCachedPlan(cachedPlanKey).Exec(statement, args...)
}

// ExecContextWithCachedPlan runs the statement without creating a QueryResult.
func (dbc *Connection) ExecContextWithCachedPlan(context context.Context, statement, cachedPlanKey string, args ...interface{}) error {
	return dbc.Invoke(context).WithCachedPlan(cachedPlanKey).Exec(statement, args...)
}

// ExecInTx runs a statement within a transaction.
func (dbc *Connection) ExecInTx(statement string, tx *sql.Tx, args ...interface{}) (err error) {
	return dbc.Invoke(dbc.Background(), tx).Exec(statement, args...)
}

// ExecInTxContext runs a statement within a transaction with a context.
func (dbc *Connection) ExecInTxContext(context context.Context, statement string, tx *sql.Tx, args ...interface{}) (err error) {
	return dbc.Invoke(context, tx).Exec(statement, args...)
}

// ExecInTxContextWithLabel runs a statement within a transaction with a label and a context.
func (dbc *Connection) ExecInTxContextWithLabel(context context.Context, statement, cachedPlanKey string, tx *sql.Tx, args ...interface{}) (err error) {
	return dbc.Invoke(context, tx).WithCachedPlan(cachedPlanKey).Exec(statement, args...)
}

// Query runs the selected statement and returns a Query.
func (dbc *Connection) Query(statement string, args ...interface{}) *Query {
	return dbc.Invoke(dbc.Background()).Query(statement, args...)
}

// QueryContext runs the selected statement and returns a Query.
func (dbc *Connection) QueryContext(context context.Context, statement string, args ...interface{}) *Query {
	return dbc.Invoke(context).Query(statement, args...)
}

// QueryWithCachedPlan runs the selected statement and returns a Query.
func (dbc *Connection) QueryWithCachedPlan(statement, cachedPlanKey string, args ...interface{}) *Query {
	return dbc.Invoke(dbc.Background()).WithCachedPlan(cachedPlanKey).Query(statement, args...)
}

// QueryContextWithCachedPlan runs the selected statement and returns a Query.
func (dbc *Connection) QueryContextWithCachedPlan(context context.Context, statement, cachedPlanKey string, args ...interface{}) *Query {
	return dbc.Invoke(context).WithCachedPlan(cachedPlanKey).Query(statement, args...)
}

// QueryInTx runs the selected statement in a transaction and returns a Query.
func (dbc *Connection) QueryInTx(statement string, tx *sql.Tx, args ...interface{}) (result *Query) {
	return dbc.Invoke(dbc.Background(), tx).Query(statement, args...)
}

// QueryInTxContext runs the selected statement in a transaction and returns a Query.
func (dbc *Connection) QueryInTxContext(context context.Context, statement string, tx *sql.Tx, args ...interface{}) (result *Query) {
	return dbc.Invoke(context, tx).Query(statement, args...)
}

// QueryInTxWithCachedPlan runs the selected statement in a transaction and returns a Query.
func (dbc *Connection) QueryInTxWithCachedPlan(statement, cachedPlanKey string, tx *sql.Tx, args ...interface{}) (result *Query) {
	return dbc.Invoke(dbc.Background(), tx).WithCachedPlan(cachedPlanKey).Query(statement, args...)
}

// QueryInTxContextWithCachedPlan runs the selected statement in a transaction and returns a Query.
func (dbc *Connection) QueryInTxContextWithCachedPlan(context context.Context, statement, QueryInTxWithCachedPlan string, tx *sql.Tx, args ...interface{}) (result *Query) {
	return dbc.Invoke(context, tx).WithCachedPlan(QueryInTxWithCachedPlan).Query(statement, args...)
}

// Get returns a given object based on a group of primary key ids.
func (dbc *Connection) Get(object DatabaseMapped, ids ...interface{}) error {
	return dbc.Invoke(dbc.Background()).Get(object, ids...)
}

// GetContext returns a given object based on a group of primary key ids using the given context.
func (dbc *Connection) GetContext(context context.Context, object DatabaseMapped, ids ...interface{}) error {
	return dbc.Invoke(context).Get(object, ids...)
}

// GetInTx returns a given object based on a group of primary key ids within a transaction.
func (dbc *Connection) GetInTx(object DatabaseMapped, tx *sql.Tx, args ...interface{}) error {
	return dbc.Invoke(dbc.Background(), tx).Get(object, args...)
}

// GetInTxContext returns a given object based on a group of primary key ids within a transaction and a given context.
func (dbc *Connection) GetInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx, args ...interface{}) error {
	return dbc.Invoke(context, tx).Get(object, args...)
}

// GetAll returns all rows of an object mapped table.
func (dbc *Connection) GetAll(collection interface{}) error {
	return dbc.Invoke(dbc.Background()).GetAll(collection)
}

// GetAllContext returns all rows of an object mapped table.
func (dbc *Connection) GetAllContext(context context.Context, collection interface{}) error {
	return dbc.Invoke(context).GetAll(collection)
}

// GetAllInTx returns all rows of an object mapped table wrapped in a transaction.
func (dbc *Connection) GetAllInTx(collection interface{}, tx *sql.Tx) error {
	return dbc.Invoke(dbc.Background(), tx).GetAll(collection)
}

// GetAllInTxContext returns all rows of an object mapped table wrapped in a transaction.
func (dbc *Connection) GetAllInTxContext(context context.Context, collection interface{}, tx *sql.Tx) error {
	return dbc.Invoke(context, tx).GetAll(collection)
}

// Create writes an object to the database.
func (dbc *Connection) Create(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).Create(object)
}

// CreateContext writes an object to the database.
func (dbc *Connection) CreateContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).Create(object)
}

// CreateInTx writes an object to the database within a transaction.
func (dbc *Connection) CreateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).Create(object)
}

// CreateInTxContext writes an object to the database within a transaction.
func (dbc *Connection) CreateInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).Create(object)
}

// CreateIfNotExists writes an object to the database if it does not already exist.
func (dbc *Connection) CreateIfNotExists(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).CreateIfNotExists(object)
}

// CreateIfNotExistsContext writes an object to the database if it does not already exist.
func (dbc *Connection) CreateIfNotExistsContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).CreateIfNotExists(object)
}

// CreateIfNotExistsInTx writes an object to the database if it does not already exist within a transaction.
func (dbc *Connection) CreateIfNotExistsInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).CreateIfNotExists(object)
}

// CreateIfNotExistsInTxContext writes an object to the database if it does not already exist within a transaction.
func (dbc *Connection) CreateIfNotExistsInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).CreateIfNotExists(object)
}

// CreateMany writes many an objects to the database.
func (dbc *Connection) CreateMany(objects interface{}) error {
	return dbc.Invoke(dbc.Background()).CreateMany(objects)
}

// CreateManyContext writes many an objects to the database.
func (dbc *Connection) CreateManyContext(context context.Context, objects interface{}) error {
	return dbc.Invoke(context).CreateMany(objects)
}

// CreateManyInTx writes many an objects to the database within a transaction.
func (dbc *Connection) CreateManyInTx(objects interface{}, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).CreateMany(objects)
}

// CreateManyInTxContext writes many an objects to the database within a transaction.
func (dbc *Connection) CreateManyInTxContext(context context.Context, objects interface{}, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).CreateMany(objects)
}

// Update updates an object.
func (dbc *Connection) Update(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).Update(object)
}

// UpdateContext updates an object.
func (dbc *Connection) UpdateContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).Update(object)
}

// UpdateInTx updates an object wrapped in a transaction.
func (dbc *Connection) UpdateInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).Update(object)
}

// UpdateInTxContext updates an object wrapped in a transaction.
func (dbc *Connection) UpdateInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).Update(object)
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbc *Connection) Exists(object DatabaseMapped) (bool, error) {
	return dbc.Invoke(dbc.Background()).Exists(object)
}

// ExistsContext returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbc *Connection) ExistsContext(context context.Context, object DatabaseMapped) (bool, error) {
	return dbc.Invoke(context).Exists(object)
}

// ExistsInTx returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbc *Connection) ExistsInTx(object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	return dbc.Invoke(dbc.Background(), tx).Exists(object)
}

// ExistsInTxContext returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbc *Connection) ExistsInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	return dbc.Invoke(context, tx).Exists(object)
}

// Delete deletes an object from the database.
func (dbc *Connection) Delete(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).Delete(object)
}

// DeleteContext deletes an object from the database.
func (dbc *Connection) DeleteContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).Delete(object)
}

// DeleteInTx deletes an object from the database wrapped in a transaction.
func (dbc *Connection) DeleteInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).Delete(object)
}

// DeleteInTxContext deletes an object from the database wrapped in a transaction.
func (dbc *Connection) DeleteInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).Delete(object)
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it.
func (dbc *Connection) Upsert(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).Upsert(object)
}

// UpsertContext inserts the object if it doesn't exist already (as defined by its primary keys) or updates it.
func (dbc *Connection) UpsertContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).Upsert(object)
}

// UpsertInTx inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (dbc *Connection) UpsertInTx(object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(dbc.Background(), tx).Upsert(object)
}

// UpsertInTxContext inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (dbc *Connection) UpsertInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) (err error) {
	return dbc.Invoke(context, tx).Upsert(object)
}

// Truncate fully removes an tables rows in a single opertation.
func (dbc *Connection) Truncate(object DatabaseMapped) error {
	return dbc.Invoke(dbc.Background()).Truncate(object)
}

// TruncateContext fully removes an tables rows in a single opertation.
func (dbc *Connection) TruncateContext(context context.Context, object DatabaseMapped) error {
	return dbc.Invoke(context).Truncate(object)
}

// TruncateInTx applies a truncation in a transaction.
func (dbc *Connection) TruncateInTx(object DatabaseMapped, tx *sql.Tx) error {
	return dbc.Invoke(dbc.Background(), tx).Truncate(object)
}

// TruncateInTxContext applies a truncation in a transaction.
func (dbc *Connection) TruncateInTxContext(context context.Context, object DatabaseMapped, tx *sql.Tx) error {
	return dbc.Invoke(context, tx).Truncate(object)
}
