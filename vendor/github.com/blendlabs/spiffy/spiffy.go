// Package spiffy providers a basic abstraction layer above normal database/sql that makes it easier to
// interact with the database and organize database related code. It is not intended to replace actual sql
// (you write queries yourself).
package spiffy

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/blendlabs/go-exception"

	// PQ is the postgres driver
	_ "github.com/lib/pq"
)

var metaCacheLock = sync.Mutex{}
var metaCache map[reflect.Type]ColumnCollection

var defaultAlias string
var defaultAliasLock = sync.Mutex{}
var dbAliasesLock = sync.Mutex{}
var dbAliases = make(map[string]*DbConnection)

// DatabaseMapped is the interface that any objects passed into database mapped methods like Create, Update, Delete, Get, GetAll etc.
// The only required method is TableName() string that returns the name of the table in the database this type is mapped to.
//
//	type MyDatabaseMappedObject {
//		Mycolumn `db:"my_column"`
//	}
//	func (_ MyDatabaseMappedObject) TableName() string {
//		return "my_database_mapped_object"
//	}
// If you require different table names based on alias, create another type.
type DatabaseMapped interface {
	TableName() string
}

// Populatable is an interface that you can implement if your object is read often and is performance critical.
type Populatable interface {
	Populate(rows *sql.Rows) error
}

// CreateDbAlias allows you to set up a connection for later use via an alias.
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//
// You can later set an alias as 'default' and refer to it using `spiffy.DefaultDb()`.
func CreateDbAlias(alias string, prototype *DbConnection) {
	dbAliasesLock.Lock()
	defer dbAliasesLock.Unlock()
	dbAliases[alias] = prototype
}

// Alias fetches a connection by its alias.
//
//	spiffy.Alias("logging").Create(&object)
//
// Alternately, if you've set the alias as 'default' you can just refer to it via. `DefaultDb()`
func Alias(alias string) *DbConnection {
	return dbAliases[alias]
}

// SetDefaultAlias sets an alias created with `CreateDbAlias` as default. This lets you refer to it later via. `DefaultDb()`
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//	spiffy.SetDefaultAlias("main")
//	execErr := spiffy.DefaultDb().Execute("select 'ok!')
//
// This will then let you refer to the alias via. `DefaultDb()`
func SetDefaultAlias(alias string) {
	defaultAliasLock.Lock()
	defer defaultAliasLock.Unlock()
	defaultAlias = alias
}

// DefaultDb returns a reference to the DbConnection set as default.
//
//	spiffy.DefaultDb().Exec("select 'ok!")
//
// Note: you must set up the default with `SetDefaultAlias()` before using DefaultDb.
func DefaultDb() *DbConnection {
	if len(defaultAlias) != 0 {
		return dbAliases[defaultAlias]
	}
	return nil
}

// NewUnauthenticatedDbConnection creates a new DbConnection without Username or Password or SSLMode
func NewUnauthenticatedDbConnection(host, schema string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = ""
	conn.Password = ""
	conn.SSLMode = "disable"
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
	return conn
}

// NewDbConnectionFromDSN creates a new connection with SSLMode set to "disable"
func NewDbConnectionFromDSN(dsn string) *DbConnection {
	conn := &DbConnection{}
	conn.DSN = dsn
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
	Tx         *sql.Tx
}

// --------------------------------------------------------------------------------
// Column
// --------------------------------------------------------------------------------

// NewColumnFromFieldTag reads the contents of a field tag, ex: `json:"foo" db:"bar,isprimarykey,isserial"
func NewColumnFromFieldTag(field reflect.StructField) *Column {
	db := field.Tag.Get("db")
	if db != "-" {
		col := Column{}
		col.FieldName = field.Name
		col.ColumnName = strings.ToLower(field.Name)
		col.FieldType = field.Type
		if db != "" {
			pieces := strings.Split(db, ",")

			if !strings.HasPrefix(db, ",") {
				col.ColumnName = pieces[0]
			}

			if len(pieces) >= 1 {
				args := strings.Join(pieces[1:], ",")
				col.IsPrimaryKey = strings.Contains(strings.ToLower(args), "pk")
				col.IsSerial = strings.Contains(strings.ToLower(args), "serial")
				col.IsNullable = strings.Contains(strings.ToLower(args), "nullable")
				col.IsReadOnly = strings.Contains(strings.ToLower(args), "readonly")
				col.IsJSON = strings.Contains(strings.ToLower(args), "json")
			}
		}
		return &col
	}

	return nil
}

// Column represents a single field on a struct that is mapped to the database.
type Column struct {
	TableName    string
	FieldName    string
	FieldType    reflect.Type
	ColumnName   string
	Index        int
	IsPrimaryKey bool
	IsSerial     bool
	IsNullable   bool
	IsReadOnly   bool
	IsJSON       bool
}

// SetValue sets the field on a database mapped object to the instance of `value`.
func (c Column) SetValue(object DatabaseMapped, value interface{}) error {
	objValue := reflectValue(object)
	field := objValue.FieldByName(c.FieldName)
	fieldType := field.Type()
	if field.CanSet() {
		valueReflected := reflectValue(value)
		if valueReflected.IsValid() {
			if c.IsJSON {
				valueAsString, ok := valueReflected.Interface().(string)
				if ok && len(valueAsString) != 0 {
					fieldAddr := field.Addr().Interface()
					jsonErr := json.Unmarshal([]byte(valueAsString), fieldAddr)
					if jsonErr != nil {
						return exception.Wrap(jsonErr)
					}
					field.Set(reflect.ValueOf(fieldAddr).Elem())
				}
			} else {
				if valueReflected.Type().AssignableTo(fieldType) {
					if field.Kind() == reflect.Ptr && valueReflected.CanAddr() {
						field.Set(valueReflected.Addr())
					} else {
						field.Set(valueReflected)
					}
				} else {
					if field.Kind() == reflect.Ptr {
						if valueReflected.CanAddr() {
							if fieldType.Elem() == valueReflected.Type() {
								field.Set(valueReflected.Addr())
							} else {
								convertedValue := valueReflected.Convert(fieldType.Elem())
								if convertedValue.CanAddr() {
									field.Set(convertedValue.Addr())
								}
							}
						}
					} else {
						convertedValue := valueReflected.Convert(fieldType)
						field.Set(convertedValue)
					}
				}
			}
		}
	} else {
		return exception.New("hit a field we can't set: '" + c.FieldName + "', did you forget to pass the object as a reference?")
	}
	return nil
}

// GetValue returns the value for a column on a given database mapped object.
func (c Column) GetValue(object DatabaseMapped) interface{} {
	value := reflectValue(object)
	valueField := value.Field(c.Index)
	return valueField.Interface()
}

// --------------------------------------------------------------------------------
// Column Collection
// --------------------------------------------------------------------------------

// NewColumnCollection creates a column lookup for a slice of columns.
func NewColumnCollection(columns []Column) ColumnCollection {
	cc := ColumnCollection{Columns: columns}
	lookup := make(map[string]*Column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.Lookup = lookup
	return cc
}

// NewColumnCollectionFromInstance reflects an object instance into a new column collection.
func NewColumnCollectionFromInstance(object DatabaseMapped) ColumnCollection {
	return NewColumnCollectionFromType(reflect.TypeOf(object))
}

// NewColumnCollectionFromType reflects a reflect.Type into a column collection.
// The results of this are cached for speed.
func NewColumnCollectionFromType(t reflect.Type) ColumnCollection {
	metaCacheLock.Lock()
	defer metaCacheLock.Unlock()

	if metaCache == nil {
		metaCache = map[reflect.Type]ColumnCollection{}
	}

	if _, ok := metaCache[t]; !ok {
		metaCache[t] = CreateColumnsByType(t)
	}
	return metaCache[t]
}

// CreateColumnsByType reflects a new column collection from a reflect.Type.
func CreateColumnsByType(t reflect.Type) ColumnCollection {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName, _ := TableName(t)
	numFields := t.NumField()

	var cols []Column
	for index := 0; index < numFields; index++ {
		field := t.Field(index)
		if !field.Anonymous {
			col := NewColumnFromFieldTag(field)
			if col != nil {
				col.Index = index
				col.TableName = tableName
				cols = append(cols, *col)
			}
		}
	}

	return NewColumnCollection(cols)
}

// ColumnCollection represents the column metadata for a given struct.
type ColumnCollection struct {
	Columns []Column
	Lookup  map[string]*Column
}

// PrimaryKeys are columns we use as where predicates and can't update.
func (cc ColumnCollection) PrimaryKeys() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if c.IsPrimaryKey {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// NotPrimaryKeys are columns we can update.
func (cc ColumnCollection) NotPrimaryKeys() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if !c.IsPrimaryKey {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// Serials are columns we have to return the id of.
func (cc ColumnCollection) Serials() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if c.IsSerial {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// NotSerials are columns we don't have to return the id of.
func (cc ColumnCollection) NotSerials() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if !c.IsSerial {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// ReadOnly are columns that we don't have to insert upon Create().
func (cc ColumnCollection) ReadOnly() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if c.IsReadOnly {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// NotReadOnly are columns that we have to insert upon Create().
func (cc ColumnCollection) NotReadOnly() ColumnCollection {
	var cols []Column
	for _, c := range cc.Columns {
		if !c.IsReadOnly {
			cols = append(cols, c)
		}
	}
	return NewColumnCollection(cols)
}

// ColumnNames returns the string names for all the columns in the collection.
func (cc ColumnCollection) ColumnNames() []string {
	var names []string
	for _, c := range cc.Columns {
		names = append(names, c.ColumnName)
	}
	return names
}

// ColumnValues returns the reflected value for all the columns on a given instance.
func (cc ColumnCollection) ColumnValues(instance interface{}) []interface{} {
	value := reflectValue(instance)

	var values []interface{}
	for _, c := range cc.Columns {
		valueField := value.FieldByName(c.FieldName)
		if c.IsJSON {
			toSerialize := valueField.Interface()
			jsonBytes, _ := json.Marshal(toSerialize)
			values = append(values, string(jsonBytes))
		} else {
			values = append(values, valueField.Interface())
		}

	}
	return values
}

// FirstOrDefault returns the first column in the collection or `nil` if the collection is empty.
func (cc ColumnCollection) FirstOrDefault() *Column {
	if len(cc.Columns) > 0 {
		col := cc.Columns[0]
		return &col
	}
	return nil
}

// ConcatWith merges a collection with another collection.
func (cc ColumnCollection) ConcatWith(other ColumnCollection) ColumnCollection {
	var total []Column
	total = append(total, cc.Columns...)
	total = append(total, other.Columns...)
	return NewColumnCollection(total)
}

// --------------------------------------------------------------------------------
// Query Result
// --------------------------------------------------------------------------------

// QueryResult is the intermediate result of a query.
type QueryResult struct {
	Rows  *sql.Rows
	Stmt  *sql.Stmt
	Error error
}

// Close closes and releases any resources retained by the QueryResult.
func (q *QueryResult) Close() error {
	if q.Rows != nil {
		if closeErr := q.Rows.Close(); closeErr != nil {
			return exception.WrapMany(q.Error, closeErr)
		}
	}
	if q.Stmt != nil {
		if closeErr := q.Stmt.Close(); closeErr != nil {
			return exception.WrapMany(q.Error, closeErr)
		}
	}

	return exception.Wrap(q.Error)
}

// Any returns if there are any results for the query.
func (q *QueryResult) Any() (bool, error) {
	if q.Error != nil {
		return false, q.Close()
	}
	hasRows := q.Rows.Next()
	q.Error = q.Rows.Err()

	return hasRows, q.Close()
}

// None returns if there are no results for the query.
func (q *QueryResult) None() (bool, error) {
	if q.Error != nil {
		return false, q.Close()
	}
	hasNoRows := !q.Rows.Next()
	q.Error = q.Rows.Err()

	return hasNoRows, q.Close()
}

// Scan writes the results to a given set of local variables.
func (q *QueryResult) Scan(args ...interface{}) error {
	if q.Error != nil {
		return q.Close()
	}

	if q.Rows.Next() {
		err := q.Rows.Scan(args...)
		if err != nil {
			q.Error = err
			return q.Close()
		}
	}

	q.Error = q.Rows.Err()
	return q.Close()
}

// Out writes the query result to a single object via. reflection mapping.
func (q *QueryResult) Out(object DatabaseMapped) error {
	if q.Error != nil {
		return q.Close()
	}

	meta := NewColumnCollectionFromInstance(object)

	if q.Rows.Next() {
		popErr := PopulateByName(object, q.Rows, meta)
		if popErr != nil {
			q.Error = popErr
			return q.Close()
		}
	}

	q.Error = q.Rows.Err()
	return q.Close()
}

// OutMany writes the query results to a slice of objects.
func (q *QueryResult) OutMany(collection interface{}) error {
	if q.Error != nil {
		return q.Close()
	}

	sliceType := reflectType(collection)
	if sliceType.Kind() != reflect.Slice {
		return exception.New("Destination collection is not a slice.")
	}

	sliceInnerType := reflectSliceType(collection)
	collectionValue := reflectValue(collection)

	meta := NewColumnCollectionFromType(sliceInnerType)

	didSetRows := false
	for q.Rows.Next() {
		newObj, _ := MakeNew(sliceInnerType)
		popErr := PopulateByName(newObj, q.Rows, meta)
		if popErr != nil {
			q.Error = popErr
			return q.Close()
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
		didSetRows = true
	}

	if !didSetRows {
		collectionValue.Set(reflect.MakeSlice(sliceType, 0, 0))
	}
	q.Error = q.Rows.Err()
	return q.Close()
}

// --------------------------------------------------------------------------------
// DbConnection
// --------------------------------------------------------------------------------

// CreatePostgresConnectionString returns a sql connection string from a given set of DbConnection parameters.
func (dbAlias *DbConnection) CreatePostgresConnectionString() string {
	if len(dbAlias.DSN) != 0 {
		return dbAlias.DSN
	}
	sslMode := "?sslmode=disable"
	if dbAlias.SSLMode != "" {
		sslMode = fmt.Sprintf("?sslmode=%s", dbAlias.SSLMode)
	}

	if dbAlias.Username != "" {
		if dbAlias.Password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s/%s%s", dbAlias.Username, dbAlias.Password, dbAlias.Host, dbAlias.Schema, sslMode)
		}
		return fmt.Sprintf("postgres://%s@%s/%s%s", dbAlias.Username, dbAlias.Host, dbAlias.Schema, sslMode)
	}
	return fmt.Sprintf("postgres://%s/%s%s", dbAlias.Host, dbAlias.Schema, sslMode)
}

// IsolateToTransaction isolates a DbConnection, globally, to a transaction. This means that any operations called after this method will use the same transaction.
func (dbAlias *DbConnection) IsolateToTransaction(tx *sql.Tx) {
	dbAlias.Tx = tx
}

// ReleaseIsolation releases an isolation, does not commit or rollback.
func (dbAlias *DbConnection) ReleaseIsolation() {
	dbAlias.Tx = nil
}

// IsIsolatedToTransaction indicates if a connection is isolated to a transaction.
func (dbAlias *DbConnection) IsIsolatedToTransaction() bool {
	return dbAlias.Tx != nil
}

// Begin starts a new transaction.
func (dbAlias *DbConnection) Begin() (*sql.Tx, error) {
	if dbAlias == nil {
		return nil, exception.New("`dbAlias` is uninitialized, cannot continue.")
	}

	if dbAlias.Tx != nil {
		return dbAlias.Tx, nil
	} else if dbAlias.Connection != nil {
		tx, txErr := dbAlias.Connection.Begin()
		return tx, exception.Wrap(txErr)
	} else {
		dbConn, dbConnErr := dbAlias.OpenNew()
		if dbConnErr != nil {
			return nil, exception.Wrap(dbConnErr)
		}
		dbAlias.Connection = dbConn
		tx, txErr := dbAlias.Connection.Begin()
		return tx, exception.Wrap(txErr)
	}
}

// Rollback rolls a given transaction back handling cases where the connection is already isolated.
func (dbAlias *DbConnection) Rollback(tx *sql.Tx) error {
	if dbAlias == nil {
		return exception.New("`dbAlias` is uninitialized, cannot rollback.")
	}

	if dbAlias.Tx != nil {
		return nil
	}
	return tx.Rollback()
}

// Commit commits a given transaction handling cases where the connection is already isolated.'
func (dbAlias *DbConnection) Commit(tx *sql.Tx) error {
	if dbAlias == nil {
		return exception.New("`dbAlias` is uninitialized, cannot commit.")
	}

	if dbAlias.Tx != nil {
		return nil
	}
	return tx.Commit()
}

// WrapInTransaction performs the given action wrapped in a transaction. Will Commit() on success and Rollback() on a non-nil error returned.
func (dbAlias *DbConnection) WrapInTransaction(action func(*sql.Tx) error) error {
	tx, err := dbAlias.Begin()
	if err != nil {
		return exception.Wrap(err)
	}
	err = action(tx)
	if dbAlias.IsIsolatedToTransaction() {
		return exception.Wrap(err)
	}
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return exception.WrapMany(rollbackErr, err)
		}
		return exception.Wrap(err)
	} else if commitErr := tx.Commit(); commitErr != nil {
		return exception.Wrap(commitErr)
	}
	return nil
}

// Prepare prepares a new statement for the connection.
func (dbAlias *DbConnection) Prepare(statement string, tx *sql.Tx) (*sql.Stmt, error) {
	if tx == nil {
		if dbAlias == nil {
			return nil, exception.New("DbConnection is nil")
		}

		if dbAlias.Tx != nil {
			stmt, stmtErr := dbAlias.Tx.Prepare(statement)
			if stmtErr != nil {
				return nil, exception.Newf("Postgres Error: %v", stmtErr)
			}
			return stmt, nil
		}

		dbConn, dbErr := dbAlias.Open()
		if dbErr != nil {
			return nil, exception.Newf("Postgres Error: %v", dbErr)
		}
		stmt, stmtErr := dbConn.Prepare(statement)
		if stmtErr != nil {
			return nil, exception.Newf("Postgres Error: %v", stmtErr)
		}
		return stmt, nil
	}

	stmt, stmtErr := tx.Prepare(statement)
	if stmtErr != nil {
		return nil, exception.Newf("Postgres Error: %v", stmtErr)
	}
	return stmt, nil
}

// OpenNew returns a new connection object.
func (dbAlias *DbConnection) OpenNew() (*sql.DB, error) {
	dbConn, err := sql.Open("postgres", dbAlias.CreatePostgresConnectionString())
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return dbConn, nil
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbAlias *DbConnection) Open() (*sql.DB, error) {
	if dbAlias.Connection == nil {
		newConn, err := dbAlias.OpenNew()
		if err != nil {
			return nil, exception.Wrap(err)
		}
		dbAlias.Connection = newConn
	}
	return dbAlias.Connection, nil
}

// Exec runs the statement without creating a QueryResult.
func (dbAlias *DbConnection) Exec(statement string, args ...interface{}) error {
	return dbAlias.ExecInTransaction(statement, nil, args...)
}

// ExecInTransaction runs a statement within a transaction.
func (dbAlias *DbConnection) ExecInTransaction(statement string, tx *sql.Tx, args ...interface{}) error {
	stmt, stmtErr := dbAlias.Prepare(statement, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()

	if _, execErr := stmt.Exec(args...); execErr != nil {
		return exception.Wrap(execErr)
	}

	return nil
}

// Query runs the selected statement and returns a QueryResult.
func (dbAlias *DbConnection) Query(statement string, args ...interface{}) *QueryResult {
	return dbAlias.QueryInTransaction(statement, nil, args...)
}

// QueryInTransaction runs the selected statement in a transaction and returns a QueryResult.
func (dbAlias *DbConnection) QueryInTransaction(statement string, tx *sql.Tx, args ...interface{}) *QueryResult {
	result := QueryResult{}

	stmt, stmtErr := dbAlias.Prepare(statement, tx)
	if stmtErr != nil {
		result.Error = exception.Wrap(stmtErr)
		return &result
	}

	rows, queryErr := stmt.Query(args...)
	if queryErr != nil {
		if closeErr := stmt.Close(); closeErr != nil {
			result.Error = exception.WrapMany(closeErr, queryErr)
		} else {
			result.Error = exception.Wrap(queryErr)
		}
		return &result
	}
	result.Stmt = stmt
	result.Rows = rows
	return &result
}

// GetByID returns a given object based on a group of primary key ids.
func (dbAlias *DbConnection) GetByID(object DatabaseMapped, ids ...interface{}) error {
	return dbAlias.GetByIDInTransaction(object, nil, ids...)
}

// GetByIDInTransaction returns a given object based on a group of primary key ids within a transaction.
func (dbAlias *DbConnection) GetByIDInTransaction(object DatabaseMapped, tx *sql.Tx, ids ...interface{}) error {
	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := NewColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	columnNames := standardCols.ColumnNames()
	tableName := object.TableName()
	pks := standardCols.PrimaryKeys()

	if len(pks.Columns) == 0 {
		return exception.New("no primary key on object to get by.")
	}

	queryBody := fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columnNames, ","), tableName, makeWhereClause(pks, 1))

	stmt, stmtErr := dbAlias.Prepare(queryBody, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()

	rows, queryErr := stmt.Query(ids...)
	if queryErr != nil {
		return exception.Wrap(queryErr)
	}

	defer rows.Close()

	for rows.Next() {
		if popErr := PopulateInOrder(object, rows, standardCols); popErr != nil {
			return exception.Wrap(popErr)
		}
	}

	return exception.Wrap(rows.Err())
}

// GetAll returns all rows of an object mapped table.
func (dbAlias *DbConnection) GetAll(collection interface{}) error {
	return dbAlias.GetAllInTransaction(collection, nil)
}

// GetAllInTransaction returns all rows of an object mapped table wrapped in a transaction.
func (dbAlias *DbConnection) GetAllInTransaction(collection interface{}, tx *sql.Tx) error {
	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)
	meta := NewColumnCollectionFromType(t).NotReadOnly()

	columnNames := meta.ColumnNames()

	sqlStmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columnNames, ","), tableName)

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()

	rows, queryErr := stmt.Query()
	if queryErr != nil {
		return exception.Wrap(queryErr)
	}
	defer rows.Close()

	for rows.Next() {
		newObj, _ := MakeNew(t)
		popErr := PopulateInOrder(newObj, rows, meta)
		if popErr != nil {
			return exception.Wrap(popErr)
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
	}

	return exception.Wrap(rows.Err())
}

// Create writes an object to the database.
func (dbAlias *DbConnection) Create(object DatabaseMapped) error {
	return dbAlias.CreateInTransaction(object, nil)
}

// CreateInTransaction writes an object to the database within a transaction.
func (dbAlias *DbConnection) CreateInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	cols := NewColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()
	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)
	tokens := makeCsvTokens(len(writeCols.Columns))

	var sqlStmt string
	if len(serials.Columns) == 0 {
		sqlStmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(colNames, ","),
			tokens,
		)
	} else {
		serial := serials.Columns[0]
		sqlStmt = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			tableName,
			strings.Join(colNames, ","),
			tokens,
			serial.ColumnName,
		)
	}

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()

	if len(serials.Columns) == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			return execErr
		}
	} else {
		serial := serials.Columns[0]

		var id interface{}
		execErr := stmt.QueryRow(colValues...).Scan(&id)
		if execErr != nil {
			return exception.Wrap(execErr)
		}
		setErr := serial.SetValue(object, id)
		if setErr != nil {
			return exception.Wrap(setErr)
		}
	}

	return nil
}

// Update updates an object.
func (dbAlias *DbConnection) Update(object DatabaseMapped) error {
	return dbAlias.UpdateInTransaction(object, nil)
}

// UpdateInTransaction updates an object wrapped in a transaction.
func (dbAlias *DbConnection) UpdateInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	tableName := object.TableName()

	cols := NewColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials().NotPrimaryKeys()
	pks := cols.PrimaryKeys()
	allCols := writeCols.ConcatWith(pks)

	totalValues := allCols.ColumnValues(object)

	numColumns := len(writeCols.Columns)

	sqlStmt := "UPDATE " + tableName + " SET "
	for i, col := range writeCols.Columns {
		sqlStmt = sqlStmt + col.ColumnName + " = $" + strconv.Itoa(i+1)
		if i != numColumns-1 {
			sqlStmt = sqlStmt + ","
		}
	}

	whereClause := makeWhereClause(pks, numColumns+1)

	sqlStmt = sqlStmt + whereClause

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()
	_, err := stmt.Exec(totalValues...)

	if err != nil {
		return exception.Wrap(err)
	}

	return nil
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbAlias *DbConnection) Exists(object DatabaseMapped) (bool, error) {
	return dbAlias.ExistsInTransaction(object, nil)
}

// ExistsInTransaction returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbAlias *DbConnection) ExistsInTransaction(object DatabaseMapped, tx *sql.Tx) (bool, error) {
	tableName := object.TableName()
	cols := NewColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns) == 0 {
		return false, exception.New("No primary key on object.")
	}
	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("SELECT 1 FROM %s %s", tableName, whereClause)
	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		return false, exception.Wrap(stmtErr)
	}
	defer stmt.Close()
	pkValues := pks.ColumnValues(object)
	rows, queryErr := stmt.Query(pkValues...)
	defer rows.Close()
	if queryErr != nil {
		return false, exception.Wrap(queryErr)
	}

	exists := rows.Next()
	return exists, nil
}

// Delete deletes an object from the database.
func (dbAlias *DbConnection) Delete(object DatabaseMapped) error {
	return dbAlias.DeleteInTransaction(object, nil)
}

// DeleteInTransaction deletes an object from the database wrapped in a transaction.
func (dbAlias *DbConnection) DeleteInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	tableName := object.TableName()
	cols := NewColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns) == 0 {
		return exception.New("No primary key on object.")
	}

	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("DELETE FROM %s %s", tableName, whereClause)

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
	if stmtErr != nil {
		return exception.Wrap(stmtErr)
	}
	defer stmt.Close()

	pkValues := pks.ColumnValues(object)

	_, err := stmt.Exec(pkValues...)
	return exception.Wrap(err)
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// reflectValue returns the reflect.Value for an object following pointers.
func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// reflectType retruns the reflect.Type for an object following pointers.
func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

// reflectSliceType returns the inner type of a slice following pointers.
func reflectSliceType(collection interface{}) reflect.Type {
	t := reflect.TypeOf(collection)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return t
}

// makeWhereClause returns the sql `where` clause for a column collection, starting at a given index (used in sql $1 parameterization).
func makeWhereClause(pks ColumnCollection, startAt int) string {
	whereClause := " WHERE "
	for i, pk := range pks.Columns {
		whereClause = whereClause + fmt.Sprintf("%s = %s", pk.ColumnName, "$"+strconv.Itoa(i+startAt))
		if i < (len(pks.Columns) - 1) {
			whereClause = whereClause + " AND "
		}
	}

	return whereClause
}

// makeCsvTokens returns a csv token string in the form "$1,$2,$3...$N"
func makeCsvTokens(num int) string {
	str := ""
	for i := 1; i <= num; i++ {
		str = str + fmt.Sprintf("$%d", i)
		if i != num {
			str = str + ","
		}
	}
	return str
}

// TableName returns the table name for a given reflect.Type by instantiating it and calling o.TableName().
// The type must implement DatabaseMapped or an exception will be returned.
func TableName(t reflect.Type) (string, error) {
	i, err := MakeNew(t)
	if err == nil {
		return i.TableName(), nil
	}
	return "", err
}

// MakeNew returns a new instance of a database mapped type.
func MakeNew(t reflect.Type) (DatabaseMapped, error) {
	newInterface := reflect.New(t).Interface()
	if typed, isTyped := newInterface.(DatabaseMapped); isTyped {
		return typed.(DatabaseMapped), nil
	}
	return nil, exception.New("`t` does not implement DatabaseMapped.")
}

func makeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

// Populate puts the contents of a sql.Rows object into a mapped object using magic reflection.
func Populate(object DatabaseMapped, row *sql.Rows) error {
	return PopulateByName(object, row, NewColumnCollectionFromInstance(object))
}

// PopulateByName sets the values of an object from the values of a sql.Rows object using column names.
func PopulateByName(object DatabaseMapped, row *sql.Rows, cols ColumnCollection) error {
	if populatable, isPopulatable := object.(Populatable); isPopulatable {
		return populatable.Populate(row)
	}

	rowColumns, rowColumnsErr := row.Columns()

	if rowColumnsErr != nil {
		return exception.Wrap(rowColumnsErr)
	}

	var values = make([]interface{}, len(rowColumns))

	for i, name := range rowColumns {
		if col, ok := cols.Lookup[name]; ok {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				values[i] = reflect.New(reflect.PtrTo(col.FieldType)).Interface()
			}
		} else {
			var value interface{}
			values[i] = &value
		}
	}

	scanErr := row.Scan(values...)

	if scanErr != nil {
		return exception.Wrap(scanErr)
	}

	for i, v := range values {
		colName := rowColumns[i]

		if field, ok := cols.Lookup[colName]; ok {
			err := field.SetValue(object, v)
			if err != nil {
				return exception.Wrap(err)
			}
		}
	}

	return nil
}

// PopulateInOrder sets the values of an object in order from a sql.Rows object.
// Only use this method if you're certain of the column order. It is faster than populateByName.
// Optionally if your object implements Populatable this process will be skipped completely, which is even faster.
func PopulateInOrder(object DatabaseMapped, row *sql.Rows, cols ColumnCollection) error {
	if populatable, isPopulatable := object.(Populatable); isPopulatable {
		return populatable.Populate(row)
	}

	var values = make([]interface{}, len(cols.Columns))

	for i, col := range cols.Columns {
		if col.FieldType.Kind() == reflect.Ptr {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				blankPtr := reflect.New(reflect.PtrTo(col.FieldType))
				if blankPtr.CanAddr() {
					values[i] = blankPtr.Addr()
				} else {
					values[i] = blankPtr.Interface()
				}
			}
		} else {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				values[i] = reflect.New(reflect.PtrTo(col.FieldType)).Interface()
			}
		}
	}

	scanErr := row.Scan(values...)

	if scanErr != nil {
		return exception.Wrap(scanErr)
	}

	for i, v := range values {
		field := cols.Columns[i]
		err := field.SetValue(object, v)
		if err != nil {
			return exception.Wrap(err)
		}
	}

	return nil
}
