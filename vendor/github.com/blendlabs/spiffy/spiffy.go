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
var metaCache map[reflect.Type]*ColumnCollection

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

// RowsConsumer is the function signature that is called from within Each().
type RowsConsumer func(r *sql.Rows) error

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

// func NewColumnCollection() *ColumnCollection { return &ColumnCollection{lookup: map[string]*Column} }

// NewColumnCollectionWithPrefix makes a new column collection with a column prefix.
func NewColumnCollectionWithPrefix(columnPrefix string) *ColumnCollection {
	return &ColumnCollection{lookup: map[string]*Column{}, columnPrefix: columnPrefix}
}

// NewColumnCollectionFromColumns creates a column lookup for a slice of columns.
func NewColumnCollectionFromColumns(columns []Column) *ColumnCollection {
	cc := ColumnCollection{columns: columns}
	lookup := make(map[string]*Column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.lookup = lookup
	return &cc
}

// NewColumnCollectionFromInstance reflects an object instance into a new column collection.
func NewColumnCollectionFromInstance(object DatabaseMapped) *ColumnCollection {
	return NewColumnCollectionFromType(reflect.TypeOf(object))
}

// NewColumnCollectionFromType reflects a reflect.Type into a column collection.
// The results of this are cached for speed.
func NewColumnCollectionFromType(t reflect.Type) *ColumnCollection {
	metaCacheLock.Lock()
	defer metaCacheLock.Unlock()

	if metaCache == nil {
		metaCache = map[reflect.Type]*ColumnCollection{}
	}

	if _, ok := metaCache[t]; !ok {
		metaCache[t] = GenerateColumnCollectionForType(t)
	}
	return metaCache[t]
}

// GenerateColumnCollectionForType reflects a new column collection from a reflect.Type.
func GenerateColumnCollectionForType(t reflect.Type) *ColumnCollection {
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

	return NewColumnCollectionFromColumns(cols)
}

// ColumnCollection represents the column metadata for a given struct.
type ColumnCollection struct {
	columns      []Column
	lookup       map[string]*Column
	columnPrefix string
}

// Len returns the number of columns.
func (cc *ColumnCollection) Len() int {
	return len(cc.columns)
}

// WithColumnPrefix applies a column prefix to column names.
func (cc *ColumnCollection) WithColumnPrefix(prefix string) *ColumnCollection {
	cc.columnPrefix = prefix
	return cc
}

// Add adds a column.
func (cc *ColumnCollection) Add(c Column) {
	cc.columns = append(cc.columns, c)
	cc.lookup[c.ColumnName] = &c
}

// PrimaryKeys are columns we use as where predicates and can't update.
func (cc ColumnCollection) PrimaryKeys() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)
	newCC.columnPrefix = cc.columnPrefix

	for _, c := range cc.columns {
		if c.IsPrimaryKey {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotPrimaryKeys are columns we can update.
func (cc ColumnCollection) NotPrimaryKeys() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsPrimaryKey {
			newCC.Add(c)
		}
	}

	return newCC
}

// Serials are columns we have to return the id of.
func (cc ColumnCollection) Serials() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if c.IsSerial {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotSerials are columns we don't have to return the id of.
func (cc ColumnCollection) NotSerials() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsSerial {
			newCC.Add(c)
		}
	}

	return newCC
}

// ReadOnly are columns that we don't have to insert upon Create().
func (cc ColumnCollection) ReadOnly() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if c.IsReadOnly {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotReadOnly are columns that we have to insert upon Create().
func (cc ColumnCollection) NotReadOnly() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsReadOnly {
			newCC.Add(c)
		}
	}

	return newCC
}

// ColumnNames returns the string names for all the columns in the collection.
func (cc ColumnCollection) ColumnNames() []string {
	var names []string
	for _, c := range cc.columns {
		if len(cc.columnPrefix) != 0 {
			names = append(names, fmt.Sprintf("%s%s", cc.columnPrefix, c.ColumnName))
		} else {
			names = append(names, c.ColumnName)
		}
	}
	return names
}

// Columns returns the colummns
func (cc ColumnCollection) Columns() []Column {
	return cc.columns
}

// Lookup gets the column name lookup.
func (cc ColumnCollection) Lookup() map[string]*Column {
	if len(cc.columnPrefix) != 0 {
		lookup := map[string]*Column{}
		for key, value := range cc.lookup {
			lookup[fmt.Sprintf("%s%s", cc.columnPrefix, key)] = value
		}
		return lookup
	}
	return cc.lookup
}

// ColumnNamesFromAlias returns the string names for all the columns in the collection.
func (cc ColumnCollection) ColumnNamesFromAlias(tableAlias string) []string {
	var names []string
	for _, c := range cc.columns {
		if len(cc.columnPrefix) != 0 {
			names = append(names, fmt.Sprintf("%s.%s as %s%s", tableAlias, c.ColumnName, cc.columnPrefix, c.ColumnName))
		} else {
			names = append(names, fmt.Sprintf("%s.%s", tableAlias, c.ColumnName))
		}
	}
	return names
}

// ColumnValues returns the reflected value for all the columns on a given instance.
func (cc ColumnCollection) ColumnValues(instance interface{}) []interface{} {
	value := reflectValue(instance)

	var values []interface{}
	for _, c := range cc.columns {
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
	if len(cc.columns) > 0 {
		col := cc.columns[0]
		return &col
	}
	return nil
}

// ConcatWith merges a collection with another collection.
func (cc ColumnCollection) ConcatWith(other *ColumnCollection) *ColumnCollection {
	var total []Column
	total = append(total, cc.columns...)
	total = append(total, other.columns...)
	return NewColumnCollectionFromColumns(total)
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

	columnMeta := NewColumnCollectionFromInstance(object)
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

	meta := NewColumnCollectionFromType(sliceInnerType)

	v, _ := MakeNew(sliceInnerType)
	isPopulatable := IsPopulatable(v)

	var popErr error
	didSetRows := false
	for q.Rows.Next() {
		newObj, _ := MakeNew(sliceInnerType)

		if isPopulatable {
			popErr = (AsPopulatable(newObj)).Populate(q.Rows)
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
	conn.MetaLock = sync.Mutex{}
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
	conn.MetaLock = sync.Mutex{}
	return conn
}

// NewDbConnectionFromDSN creates a new connection with SSLMode set to "disable"
func NewDbConnectionFromDSN(dsn string) *DbConnection {
	conn := &DbConnection{}
	conn.DSN = dsn
	conn.MetaLock = sync.Mutex{}
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
	conn.MetaLock = sync.Mutex{}
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
	MetaLock   sync.Mutex
}

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

// Begin starts a new transaction.
func (dbAlias *DbConnection) Begin() (*sql.Tx, error) {
	if dbAlias == nil {
		return nil, exception.New("`dbAlias` is uninitialized, cannot continue.")
	}

	if dbAlias.Connection != nil {
		tx, txErr := dbAlias.Connection.Begin()
		return tx, exception.Wrap(txErr)
	}

	connection, err := dbAlias.Open()
	if err != nil {
		return nil, exception.Wrap(err)
	}
	tx, err := connection.Begin()
	return tx, exception.Wrap(err)
}

// WrapInTransaction performs the given action wrapped in a transaction. Will Commit() on success and Rollback() on a non-nil error returned.
func (dbAlias *DbConnection) WrapInTransaction(action func(*sql.Tx) error) error {
	tx, err := dbAlias.Begin()
	if err != nil {
		return exception.Wrap(err)
	}
	err = action(tx)
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
	if dbAlias == nil {
		return nil, exception.New("DbConnection is nil")
	}

	if tx != nil {
		stmt, stmtErr := tx.Prepare(statement)
		if stmtErr != nil {
			return nil, exception.Newf("Postgres Error: %v", stmtErr)
		}
		return stmt, nil
	}

	// open shared connection
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
		dbAlias.MetaLock.Lock()
		defer dbAlias.MetaLock.Unlock()

		if dbAlias.Connection == nil {
			newConn, err := dbAlias.OpenNew()
			if err != nil {
				return nil, exception.Wrap(err)
			}
			dbAlias.Connection = newConn
		}
	}
	return dbAlias.Connection, nil
}

// Exec runs the statement without creating a QueryResult.
func (dbAlias *DbConnection) Exec(statement string, args ...interface{}) error {
	return dbAlias.ExecInTransaction(statement, nil, args...)
}

// ExecInTransaction runs a statement within a transaction.
func (dbAlias *DbConnection) ExecInTransaction(statement string, tx *sql.Tx, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	stmt, stmtErr := dbAlias.Prepare(statement, tx)
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
func (dbAlias *DbConnection) Query(statement string, args ...interface{}) *QueryResult {
	return dbAlias.QueryInTransaction(statement, nil, args...)
}

// QueryInTransaction runs the selected statement in a transaction and returns a QueryResult.
func (dbAlias *DbConnection) QueryInTransaction(statement string, tx *sql.Tx, args ...interface{}) (result *QueryResult) {
	result = &QueryResult{}

	stmt, stmtErr := dbAlias.Prepare(statement, tx)
	if stmtErr != nil {
		result.Error = exception.Wrap(stmtErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			closeErr := stmt.Close()
			result.Error = exception.WrapMany(result.Error, exception.New(r), closeErr)
		}
	}()

	rows, queryErr := stmt.Query(args...)
	if queryErr != nil {
		result.Error = exception.Wrap(queryErr)
		return
	}

	// the result MUST close these.
	result.Stmt = stmt
	result.Rows = rows
	return
}

// GetByID returns a given object based on a group of primary key ids.
func (dbAlias *DbConnection) GetByID(object DatabaseMapped, ids ...interface{}) error {
	return dbAlias.GetByIDInTransaction(object, nil, ids...)
}

// GetByIDInTransaction returns a given object based on a group of primary key ids within a transaction.
func (dbAlias *DbConnection) GetByIDInTransaction(object DatabaseMapped, tx *sql.Tx, ids ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := NewColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	columnNames := standardCols.ColumnNames()
	tableName := object.TableName()
	pks := standardCols.PrimaryKeys()

	if pks.Len() == 0 {
		err = exception.New("no primary key on object to get by.")
		return
	}

	queryBody := fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(columnNames, ","), tableName, makeWhereClause(pks, 1))

	stmt, stmtErr := dbAlias.Prepare(queryBody, tx)
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
			popErr = (AsPopulatable(object)).Populate(rows)
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
func (dbAlias *DbConnection) GetAll(collection interface{}) error {
	return dbAlias.GetAllInTransaction(collection, nil)
}

// GetAllInTransaction returns all rows of an object mapped table wrapped in a transaction.
func (dbAlias *DbConnection) GetAllInTransaction(collection interface{}, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)
	meta := NewColumnCollectionFromType(t).NotReadOnly()

	columnNames := meta.ColumnNames()
	sqlStmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columnNames, ","), tableName)

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
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
			popErr = (AsPopulatable(newObj)).Populate(rows)
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
func (dbAlias *DbConnection) Create(object DatabaseMapped) error {
	return dbAlias.CreateInTransaction(object, nil)
}

// CreateInTransaction writes an object to the database within a transaction.
func (dbAlias *DbConnection) CreateInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	cols := NewColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()
	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)
	tokens := makeCsvTokens(writeCols.Len())

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

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
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
func (dbAlias *DbConnection) Update(object DatabaseMapped) error {
	return dbAlias.UpdateInTransaction(object, nil)
}

// UpdateInTransaction updates an object wrapped in a transaction.
func (dbAlias *DbConnection) UpdateInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	tableName := object.TableName()
	cols := NewColumnCollectionFromInstance(object)
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

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
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
		err = exception.Wrap(err)
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist).
func (dbAlias *DbConnection) Exists(object DatabaseMapped) (bool, error) {
	return dbAlias.ExistsInTransaction(object, nil)
}

// ExistsInTransaction returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (dbAlias *DbConnection) ExistsInTransaction(object DatabaseMapped, tx *sql.Tx) (exists bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	tableName := object.TableName()
	cols := NewColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if pks.Len() == 0 {
		exists = false
		err = exception.New("No primary key on object.")
		return
	}
	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("SELECT 1 FROM %s %s", tableName, whereClause)
	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
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
func (dbAlias *DbConnection) Delete(object DatabaseMapped) error {
	return dbAlias.DeleteInTransaction(object, nil)
}

// DeleteInTransaction deletes an object from the database wrapped in a transaction.
func (dbAlias *DbConnection) DeleteInTransaction(object DatabaseMapped, tx *sql.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			recoveryException := exception.New(r)
			err = exception.WrapMany(err, recoveryException)
		}
	}()

	tableName := object.TableName()
	cols := NewColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	whereClause := makeWhereClause(pks, 1)
	sqlStmt := fmt.Sprintf("DELETE FROM %s %s", tableName, whereClause)

	stmt, stmtErr := dbAlias.Prepare(sqlStmt, tx)
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

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// AsPopulatable casts an object as populatable.
func AsPopulatable(object DatabaseMapped) Populatable {
	return object.(Populatable)
}

// IsPopulatable returns if an object is populatable
func IsPopulatable(object DatabaseMapped) bool {
	_, isPopulatable := object.(Populatable)
	return isPopulatable
}

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
func makeWhereClause(pks *ColumnCollection, startAt int) string {
	whereClause := " WHERE "
	for i, pk := range pks.Columns() {
		whereClause = whereClause + fmt.Sprintf("%s = %s", pk.ColumnName, "$"+strconv.Itoa(i+startAt))
		if i < (pks.Len() - 1) {
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

// PopulateByName sets the values of an object from the values of a sql.Rows object using column names.
func PopulateByName(object DatabaseMapped, row *sql.Rows, cols *ColumnCollection) error {
	rowColumns, rowColumnsErr := row.Columns()

	if rowColumnsErr != nil {
		return exception.Wrap(rowColumnsErr)
	}

	var values = make([]interface{}, len(rowColumns))
	var columnLookup = cols.Lookup()

	for i, name := range rowColumns {
		if col, ok := columnLookup[name]; ok {
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

		if field, ok := columnLookup[colName]; ok {
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
func PopulateInOrder(object DatabaseMapped, row *sql.Rows, cols *ColumnCollection) error {
	var values = make([]interface{}, cols.Len())

	for i, col := range cols.Columns() {
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

	columns := cols.Columns()
	for i, v := range values {
		field := columns[i]
		err := field.SetValue(object, v)
		if err != nil {
			return exception.Wrap(err)
		}
	}

	return nil
}
