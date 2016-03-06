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
	_ "github.com/lib/pq"
)

var metaCacheLock sync.Mutex = sync.Mutex{}
var metaCache map[reflect.Type]columnCollection

var defaultAlias string
var defaultAliasLock sync.Mutex = sync.Mutex{}
var dbAliasesLock sync.Mutex = sync.Mutex{}
var dbAliases map[string]*DbConnection = make(map[string]*DbConnection)

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

/// Populatable is an interface that you can implement if your object is read often and is performance critical
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

// Fetch a connection by its alias.
//
//	spiffy.Alias("logging").Create(&object)
//
// Alternately, if you've set the alias as 'default' you can just refer to it via. `DefaultDb()`
func Alias(alias string) *DbConnection {
	return dbAliases[alias]
}

// Sets an alias created with `CreateDbAlias` as default. This lets you refer to it later via. `DefaultDb()`
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

// Returns a reference to the DbConnection set as default.
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

// Creates a new DbConnection without Username or Password or SSLMode
func NewUnauthenticatedDbConnection(host, schema string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = ""
	conn.Password = ""
	conn.SSLMode = "disable"
	return conn
}

// Creates a new connection with SSLMode set to "disable"
func NewDbConnection(host, schema, username, password string) *DbConnection {
	conn := &DbConnection{}
	conn.Host = host
	conn.Schema = schema
	conn.Username = username
	conn.Password = password
	conn.SSLMode = "disable"
	return conn
}

// Creates a new connection with SSLMode set to "disable"
func NewDbConnectionFromDSN(dsn string) *DbConnection {
	conn := &DbConnection{}
	conn.DSN = dsn
	return conn
}

// Creates a new connection with all available options (including SSLMode)
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

type column struct {
	TableName    string
	FieldName    string
	FieldType    reflect.Type
	ColumnName   string
	Index        int
	IsPrimaryKey bool
	IsSerial     bool
	IsNullable   bool
	IsReadOnly   bool
	IsJson       bool
}

func (c column) SetValue(object DatabaseMapped, value interface{}) error {
	objValue := reflectValue(object)
	field := objValue.FieldByName(c.FieldName)
	fieldType := field.Type()
	if field.CanSet() {
		valueReflected := reflectValue(value)
		if valueReflected.IsValid() {
			if c.IsJson {
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

func (c column) GetValue(object DatabaseMapped) interface{} {
	value := reflectValue(object)
	valueField := value.Field(c.Index)
	return valueField.Interface()
}

// --------------------------------------------------------------------------------
// Column Collection
// --------------------------------------------------------------------------------

type columnCollection struct {
	Columns []column
	Lookup  map[string]*column
}

func newColumnCollection(columns []column) columnCollection {
	cc := columnCollection{Columns: columns}
	lookup := make(map[string]*column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.Lookup = lookup
	return cc
}

//things we use as where predicates and can't update
func (cc columnCollection) PrimaryKeys() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if c.IsPrimaryKey {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

//things we can update
func (cc columnCollection) NotPrimaryKeys() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if !c.IsPrimaryKey {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

//things we have to return the id of ...
func (cc columnCollection) Serials() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if c.IsSerial {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

//things we don't have to return the id of ...
func (cc columnCollection) NotSerials() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if !c.IsSerial {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

//a.k.a. not things we insert
func (cc columnCollection) ReadOnly() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if c.IsReadOnly {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

func (cc columnCollection) NotReadonly() columnCollection {
	var cols []column
	for _, c := range cc.Columns {
		if !c.IsReadOnly {
			cols = append(cols, c)
		}
	}
	return newColumnCollection(cols)
}

func (cc columnCollection) ColumnNames() []string {
	var names []string
	for _, c := range cc.Columns {
		names = append(names, c.ColumnName)
	}
	return names
}

func (cc columnCollection) ColumnValues(instance interface{}) []interface{} {
	value := reflectValue(instance)

	var values []interface{}
	for _, c := range cc.Columns {
		valueField := value.FieldByName(c.FieldName)
		if c.IsJson {
			toSerialize := valueField.Interface()
			jsonBytes, _ := json.Marshal(toSerialize)
			values = append(values, string(jsonBytes))
		} else {
			values = append(values, valueField.Interface())
		}

	}
	return values
}

func (cc columnCollection) FirstOrDefault() *column {
	if len(cc.Columns) > 0 {
		col := cc.Columns[0]
		return &col
	}
	return nil
}

func (cc columnCollection) ConcatWith(other columnCollection) columnCollection {
	var total []column
	total = append(total, cc.Columns...)
	total = append(total, other.Columns...)
	return newColumnCollection(total)
}

// --------------------------------------------------------------------------------
// Query Result
// --------------------------------------------------------------------------------

type queryResult struct {
	Rows  *sql.Rows
	Stmt  *sql.Stmt
	Error error
}

func (q *queryResult) cleanup() error {
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

func (q *queryResult) Scan(args ...interface{}) error {
	if q.Error != nil {
		return q.cleanup()
	}

	if q.Rows.Next() {
		err := q.Rows.Scan(args...)
		if err != nil {
			q.Error = err
			return q.cleanup()
		}
	}

	q.Error = q.Rows.Err()
	return q.cleanup()
}

func (q *queryResult) Out(object DatabaseMapped) error {
	if q.Error != nil {
		return q.cleanup()
	}

	meta := getColumns(object)

	if q.Rows.Next() {
		popErr := populateByName(object, q.Rows, meta)
		if popErr != nil {
			q.Error = popErr
			return q.cleanup()
		}
	}

	q.Error = q.Rows.Err()
	return q.cleanup()
}

func (q *queryResult) OutMany(collection interface{}) error {
	if q.Error != nil {
		return q.cleanup()
	}

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	sliceType := reflectType(collection)
	meta := getColumnsByType(t)

	didSetRows := false
	for q.Rows.Next() {
		newObj := makeNew(t)
		popErr := populateByName(newObj, q.Rows, meta)
		if popErr != nil {
			q.Error = popErr
			return q.cleanup()
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
		didSetRows = true
	}

	if !didSetRows {
		collectionValue.Set(reflect.MakeSlice(sliceType, 0, 0))
	}
	q.Error = q.Rows.Err()
	return q.cleanup()
}

// --------------------------------------------------------------------------------
// DbConnection
// --------------------------------------------------------------------------------

// Returns a sql connection string from a given set of DbConnection parameters
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
		} else {
			return fmt.Sprintf("postgres://%s@%s/%s%s", dbAlias.Username, dbAlias.Host, dbAlias.Schema, sslMode)
		}
	} else {
		return fmt.Sprintf("postgres://%s/%s%s", dbAlias.Host, dbAlias.Schema, sslMode)
	}
}

// Isolates a DbConnection, globally, to a transaction. This means that any operations called after this method will use the same transaction.
func (dbAlias *DbConnection) IsolateToTransaction(tx *sql.Tx) {
	dbAlias.Tx = tx
}

// Releases an isolation, does not commit or rollback.
func (dbAlias *DbConnection) ReleaseIsolation() {
	dbAlias.Tx = nil
}

// Indicates if a connection is isolated to a transaction.
func (dbAlias *DbConnection) IsIsolatedToTransaction() bool {
	return dbAlias.Tx != nil
}

// Starts a new transaction.
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

// Performs the given action wrapped in a transaction. Will Commit() on success and Rollback() on a non-nil error returned.
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
		} else {
			return exception.Wrap(err)
		}
	} else if commitErr := tx.Commit(); commitErr != nil {
		return exception.Wrap(commitErr)
	}
	return nil
}

// Prepares a new statement for the connection.
func (dbAlias *DbConnection) Prepare(statement string, tx *sql.Tx) (*sql.Stmt, error) {
	if tx == nil {
		if dbAlias == nil {
			return nil, exception.New("DbConnection is nil")
		}

		if dbAlias.Tx != nil {
			if stmt, stmtErr := dbAlias.Tx.Prepare(statement); stmtErr != nil {
				return nil, exception.Newf("Postgres Error: %v", stmtErr)
			} else {
				return stmt, nil
			}
		} else {
			if dbConn, dbErr := dbAlias.Open(); dbErr != nil {
				return nil, exception.Newf("Postgres Error: %v", dbErr)
			} else {
				if stmt, stmtErr := dbConn.Prepare(statement); stmtErr != nil {
					return nil, exception.Newf("Postgres Error: %v", stmtErr)
				} else {
					return stmt, nil
				}
			}
		}
	} else {
		if stmt, stmtErr := tx.Prepare(statement); stmtErr != nil {
			return nil, exception.Newf("Postgres Error: %v", stmtErr)
		} else {
			return stmt, nil
		}

	}
}

func (dbAlias *DbConnection) OpenNew() (*sql.DB, error) {
	if dbConn, err := sql.Open("postgres", dbAlias.CreatePostgresConnectionString()); err != nil {
		return nil, exception.Wrap(err)
	} else {
		return dbConn, nil
	}
}

func (dbAlias *DbConnection) Open() (*sql.DB, error) {
	if dbAlias.Connection == nil {
		if newConn, err := dbAlias.OpenNew(); err != nil {
			return nil, exception.Wrap(err)
		} else {
			dbAlias.Connection = newConn
		}
	}
	return dbAlias.Connection, nil
}

func (dbAlias *DbConnection) Exec(statement string, args ...interface{}) error {
	return dbAlias.ExecInTransaction(statement, nil, args...)
}

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

func (dbAlias *DbConnection) Query(statement string, args ...interface{}) *queryResult {
	return dbAlias.QueryInTransaction(statement, nil, args...)
}

func (dbAlias *DbConnection) QueryInTransaction(statement string, tx *sql.Tx, args ...interface{}) *queryResult {
	result := queryResult{}

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

func (dbAlias *DbConnection) GetById(object DatabaseMapped, ids ...interface{}) error {
	return dbAlias.GetByIdInTransaction(object, nil, ids...)
}

func (dbAlias *DbConnection) GetByIdInTransaction(object DatabaseMapped, tx *sql.Tx, ids ...interface{}) error {
	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := getColumns(object)
	standardCols := meta.NotReadonly()
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
		if popErr := populateInOrder(object, rows, standardCols); popErr != nil {
			return exception.Wrap(popErr)
		}
	}

	return exception.Wrap(rows.Err())
}

func (dbAlias *DbConnection) GetAll(collection interface{}) error {
	return dbAlias.GetAllInTransaction(collection, nil)
}

func (dbAlias *DbConnection) GetAllInTransaction(collection interface{}, tx *sql.Tx) error {
	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName := tableName(t)
	meta := getColumnsByType(t).NotReadonly()

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
		newObj := makeNew(t)
		popErr := populateInOrder(newObj, rows, meta)
		if popErr != nil {
			return exception.Wrap(popErr)
		}
		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
	}

	return exception.Wrap(rows.Err())
}

func (dbAlias *DbConnection) Create(object DatabaseMapped) error {
	return dbAlias.CreateInTransaction(object, nil)
}

func (dbAlias *DbConnection) CreateInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	cols := getColumns(object)
	writeCols := cols.NotReadonly().NotSerials()

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

func (dbAlias *DbConnection) Update(object DatabaseMapped) error {
	return dbAlias.UpdateInTransaction(object, nil)
}

func (dbAlias *DbConnection) UpdateInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	tableName := object.TableName()

	cols := getColumns(object)
	writeCols := cols.NotReadonly().NotSerials().NotPrimaryKeys()
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

func (dbAlias *DbConnection) Exists(object DatabaseMapped) (bool, error) {
	return dbAlias.ExistsInTransaction(object, nil)
}

func (dbAlias *DbConnection) ExistsInTransaction(object DatabaseMapped, tx *sql.Tx) (bool, error) {
	tableName := object.TableName()
	cols := getColumns(object)
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

func (dbAlias *DbConnection) Delete(object DatabaseMapped) error {
	return dbAlias.DeleteInTransaction(object, nil)
}

func (dbAlias *DbConnection) DeleteInTransaction(object DatabaseMapped, tx *sql.Tx) error {
	tableName := object.TableName()
	cols := getColumns(object)
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

func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

func reflectSliceType(collection interface{}) reflect.Type {
	t := reflect.TypeOf(collection)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return t
}

func makeWhereClause(pks columnCollection, start_at int) string {
	whereClause := " WHERE "
	for i, pk := range pks.Columns {
		whereClause = whereClause + fmt.Sprintf("%s = %s", pk.ColumnName, "$"+strconv.Itoa(i+start_at))
		if i < (len(pks.Columns) - 1) {
			whereClause = whereClause + " AND "
		}
	}

	return whereClause
}

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

func tableName(t reflect.Type) string {
	return makeNew(t).TableName()
}

func makeNew(t reflect.Type) DatabaseMapped {
	newInterface := reflect.New(t).Interface()
	return newInterface.(DatabaseMapped)
}

func makeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

func populate(object DatabaseMapped, row *sql.Rows) error {
	return populateByName(object, row, getColumns(object))
}

func populateByName(object DatabaseMapped, row *sql.Rows, cols columnCollection) error {
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
			if col.IsJson {
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

func populateInOrder(object DatabaseMapped, row *sql.Rows, cols columnCollection) error {

	if populatable, isPopulatable := object.(Populatable); isPopulatable {
		return populatable.Populate(row)
	}

	var values = make([]interface{}, len(cols.Columns))

	for i, col := range cols.Columns {
		if col.FieldType.Kind() == reflect.Ptr {
			if col.IsJson {
				str := ""
				values[i] = &str
			} else {
				blank_ptr := reflect.New(reflect.PtrTo(col.FieldType))
				if blank_ptr.CanAddr() {
					values[i] = blank_ptr.Addr()
				} else {
					values[i] = blank_ptr.Interface()
				}
			}
		} else {
			if col.IsJson {
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

func getColumns(object DatabaseMapped) columnCollection {
	return getColumnsByType(reflect.TypeOf(object))
}

func getColumnsByType(t reflect.Type) columnCollection {
	metaCacheLock.Lock()
	defer metaCacheLock.Unlock()

	if metaCache == nil {
		metaCache = map[reflect.Type]columnCollection{}
	}

	if _, ok := metaCache[t]; !ok {
		metaCache[t] = createColumnsByType(t)
	}
	return metaCache[t]
}

func createColumnsByType(t reflect.Type) columnCollection {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName := tableName(t)
	numFields := t.NumField()

	var cols []column
	for index := 0; index < numFields; index++ {
		field := t.Field(index)
		if !field.Anonymous {
			col := readFieldTag(field)
			if col != nil {
				col.Index = index
				col.TableName = tableName
				cols = append(cols, *col)
			}
		}
	}

	return newColumnCollection(cols)
}

// reads the contents of a field tag, ex: `json:"foo" db:"bar,isprimarykey,isserial"
func readFieldTag(field reflect.StructField) *column {
	db := field.Tag.Get("db")
	if db != "-" {
		col := column{}
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
				col.IsJson = strings.Contains(strings.ToLower(args), "json")
			}
		}
		return &col
	}

	return nil
}
