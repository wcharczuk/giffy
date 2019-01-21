package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// Invocation is a specific operation against a context.
type Invocation struct {
	cachedPlanKey string

	conn                 *Connection
	context              context.Context
	cancel               func()
	statementInterceptor StatementInterceptor
	tracer               Tracer
	traceFinisher        TraceFinisher
	startTime            time.Time
	tx                   *sql.Tx
}

// StartTime returns the invocation start time.
func (i *Invocation) StartTime() time.Time {
	return i.startTime
}

// WithContext sets the context and returns a reference to the invocation.
func (i *Invocation) WithContext(context context.Context) *Invocation {
	i.context = context
	return i
}

// Context returns the underlying context.
func (i *Invocation) Context() context.Context {
	if i.context == nil {
		return context.Background()
	}
	return i.context
}

// WithCancel sets an optional cancel callback.
func (i *Invocation) WithCancel(cancel func()) *Invocation {
	i.cancel = cancel
	return i
}

// Cancel returns the optional cancel callback.
func (i *Invocation) Cancel() func() {
	return i.cancel
}

// WithCachedPlan instructs the query generator to get or create a cached prepared statement.
func (i *Invocation) WithCachedPlan(cacheKey string) *Invocation {
	i.cachedPlanKey = cacheKey
	return i
}

// CachedPlanKey returns the statement / plan cache label for the context.
func (i *Invocation) CachedPlanKey() string {
	return i.cachedPlanKey
}

// WithTx sets the tx
func (i *Invocation) WithTx(tx *sql.Tx) *Invocation {
	i.tx = tx
	return i
}

// Tx returns the underlying transaction.
func (i *Invocation) Tx() *sql.Tx {
	return i.tx
}

// WithStatementInterceptor sets the connection statement interceptor.
func (i *Invocation) WithStatementInterceptor(interceptor StatementInterceptor) *Invocation {
	i.statementInterceptor = interceptor
	return i
}

// StatementInterceptor returns the statement interceptor.
func (i *Invocation) StatementInterceptor() StatementInterceptor {
	return i.statementInterceptor
}

// Prepare returns a cached or newly prepared statment plan for a given sql statement.
func (i *Invocation) Prepare(statement string) (stmt *sql.Stmt, err error) {
	if i.statementInterceptor != nil {
		statement, err = i.statementInterceptor(i.cachedPlanKey, statement)
		if err != nil {
			return
		}
	}
	stmt, err = i.conn.PrepareContext(i.Context(), i.cachedPlanKey, statement, i.tx)
	return
}

// Exec executes a sql statement with a given set of arguments.
func (i *Invocation) Exec(statement string, args ...interface{}) (err error) {
	var stmt *sql.Stmt
	statement, err = i.start(statement)
	defer func() { err = i.finish(statement, recover(), err) }()
	if err != nil {
		return
	}

	stmt, err = i.Prepare(statement)
	if err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if _, err = stmt.ExecContext(i.Context(), args...); err != nil {
		err = Error(err)
		return
	}
	return
}

// Query returns a new query object for a given sql query and arguments.
func (i *Invocation) Query(statement string, args ...interface{}) *Query {
	var err error
	statement, err = i.start(statement)
	return &Query{
		context:       i.Context(),
		statement:     statement,
		cachedPlanKey: i.cachedPlanKey,
		args:          args,
		conn:          i.conn,
		inv:           i,
		tx:            i.tx,
		err:           err,
	}
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (i *Invocation) Get(object DatabaseMapped, ids ...interface{}) (err error) {
	if len(ids) == 0 {
		err = Error(ErrInvalidIDs)
		return
	}

	var queryBody string
	var stmt *sql.Stmt
	var cols *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), err) }()

	if i.cachedPlanKey, queryBody, cols, err = i.generateGet(object); err != nil {
		err = Error(err)
		return
	}

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = exception.New(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	row := stmt.QueryRowContext(i.Context(), ids...)
	var populateErr error
	if typed, ok := object.(Populatable); ok {
		populateErr = typed.Populate(row)
	} else {
		populateErr = PopulateInOrder(object, row, cols)
	}
	if populateErr != nil && !exception.Is(populateErr, sql.ErrNoRows) {
		err = Error(populateErr)
		return
	}

	return
}

// GetAll returns all rows of an object mapped table wrapped in a transaction.
func (i *Invocation) GetAll(collection interface{}) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	var rows *sql.Rows
	var cols *ColumnCollection
	var collectionType reflect.Type
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody, cols, collectionType = i.generateGetAll(collection)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if rows, err = stmt.QueryContext(i.Context()); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = exception.Nest(err, rows.Close()) }()

	collectionValue := reflectValue(collection)
	for rows.Next() {
		var obj interface{}
		if obj, err = makeNewDatabaseMapped(collectionType); err != nil {
			err = exception.New(err)
			return
		}

		if typed, ok := obj.(Populatable); ok {
			err = typed.Populate(rows)
		} else {
			err = PopulateInOrder(obj, rows, cols)
		}
		if err != nil {
			err = Error(err)
			return
		}

		objValue := reflectValue(obj)
		collectionValue.Set(reflect.Append(collectionValue, objValue))
	}
	return
}

// Create writes an object to the database within a transaction.
func (i *Invocation) Create(object DatabaseMapped) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	var writeCols, autos *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody, writeCols, autos = i.generateCreate(object)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if autos.Len() == 0 {
		if _, err = stmt.ExecContext(i.Context(), writeCols.ColumnValues(object)...); err != nil {
			err = Error(err)
			return
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = stmt.QueryRowContext(i.Context(), writeCols.ColumnValues(object)...).Scan(autoValues...); err != nil {
		err = Error(err)
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		err = Error(err)
		return
	}

	return
}

// CreateIfNotExists writes an object to the database if it does not already exist within a transaction.
func (i *Invocation) CreateIfNotExists(object DatabaseMapped) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	var autos, writeCols *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody, autos, writeCols = i.generateCreateIfNotExists(object)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if autos.Len() == 0 {
		if _, err = stmt.ExecContext(i.context, writeCols.ColumnValues(object)...); err != nil {
			err = Error(err)
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = stmt.QueryRowContext(i.Context(), writeCols.ColumnValues(object)...).Scan(autoValues...); err != nil {
		err = Error(err)
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		err = Error(err)
		return
	}

	return
}

// CreateMany writes many objects to the database in a single insert.
// Important; this will not use cached statements ever because the generated query
// is different for each cardinality of objects.
func (i *Invocation) CreateMany(objects interface{}) (err error) {
	var queryBody string
	var writeCols *ColumnCollection
	var sliceValue reflect.Value
	defer func() { err = i.finish(queryBody, recover(), err) }()

	queryBody, writeCols, sliceValue = i.generateCreateMany(objects)
	if sliceValue.Len() == 0 {
		// If there is nothing to create, then we're done here
		return
	}

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}

	var colValues []interface{}
	for row := 0; row < sliceValue.Len(); row++ {
		colValues = append(colValues, writeCols.ColumnValues(sliceValue.Index(row).Interface())...)
	}

	if i.tx != nil {
		_, err = i.tx.ExecContext(i.Context(), queryBody, colValues...)
	} else {
		_, err = i.conn.connection.ExecContext(i.Context(), queryBody, colValues...)
	}
	if err != nil {
		err = Error(err)
		return
	}
	return
}

// Update updates an object wrapped in a transaction.
func (i *Invocation) Update(object DatabaseMapped) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	var pks, writeCols *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody, pks, writeCols = i.generateUpdate(object)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if _, err = stmt.ExecContext(i.Context(), append(writeCols.ColumnValues(object), pks.ColumnValues(object)...)...); err != nil {
		err = Error(err)
		return
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (i *Invocation) Upsert(object DatabaseMapped) (err error) {
	var queryBody string
	var autos, writeCols *ColumnCollection
	var stmt *sql.Stmt
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody, autos, writeCols = i.generateUpsert(object)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if autos.Len() == 0 {
		if _, err = stmt.ExecContext(i.Context(), writeCols.ColumnValues(object)...); err != nil {
			err = Error(err)
			return
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = stmt.QueryRowContext(i.Context(), writeCols.ColumnValues(object)...).Scan(autoValues...); err != nil {
		err = Error(err)
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		err = Error(err)
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (i *Invocation) Exists(object DatabaseMapped) (exists bool, err error) {
	var queryBody string
	var pks *ColumnCollection
	var stmt *sql.Stmt
	defer func() { err = i.finish(queryBody, recover(), err) }()

	if i.cachedPlanKey, queryBody, pks, err = i.generateExists(object); err != nil {
		err = Error(err)
		return
	}
	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = exception.New(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	var value int
	if queryErr := stmt.QueryRowContext(i.Context(), pks.ColumnValues(object)...).Scan(&value); queryErr != nil && !exception.Is(queryErr, sql.ErrNoRows) {
		err = Error(queryErr)
		return
	}

	exists = value != 0
	return
}

// Delete deletes an object from the database wrapped in a transaction.
func (i *Invocation) Delete(object DatabaseMapped) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	var pks *ColumnCollection
	defer func() { err = i.finish(queryBody, recover(), err) }()

	if i.cachedPlanKey, queryBody, pks, err = i.generateDelete(object); err != nil {
		return
	}

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if _, err = stmt.ExecContext(i.Context(), pks.ColumnValues(object)...); err != nil {
		err = Error(err)
		return
	}
	return
}

// Truncate completely empties a table in a single command.
func (i *Invocation) Truncate(object DatabaseMapped) (err error) {
	var queryBody string
	var stmt *sql.Stmt
	defer func() { err = i.finish(queryBody, recover(), err) }()

	i.cachedPlanKey, queryBody = i.generateTruncate(object)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if stmt, err = i.Prepare(queryBody); err != nil {
		err = Error(err)
		return
	}
	defer func() { err = i.closeStatement(stmt, err) }()

	if _, err = stmt.ExecContext(i.Context()); err != nil {
		err = Error(err)
		return
	}
	return
}

// --------------------------------------------------------------------------------
// query body generators
// --------------------------------------------------------------------------------

func (i *Invocation) generateGet(object DatabaseMapped) (statementLabel, queryBody string, cols *ColumnCollection, err error) {
	tableName := TableName(object)

	cols = getCachedColumnCollectionFromInstance(object).NotReadOnly()
	pks := cols.PrimaryKeys()
	if pks.Len() == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range cols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}

	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")

	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}

	statementLabel = tableName + "_get"
	queryBody = queryBodyBuffer.String()
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateGetAll(collection interface{}) (statementLabel, queryBody string, cols *ColumnCollection, collectionType reflect.Type) {
	collectionType = reflectSliceType(collection)
	tableName := TableNameByType(collectionType)

	cols = getCachedColumnCollectionFromType(tableName, reflectSliceType(collection)).NotReadOnly()

	queryBodyBuffer := i.conn.bufferPool.Get()
	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range cols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_get_all"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateCreate(object DatabaseMapped) (statementLabel, queryBody string, writeCols, autos *ColumnCollection) {
	tableName := TableName(object)

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols = cols.WriteColumns()
	autos = cols.Autos()

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range writeCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < writeCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(")")

	if autos.Len() > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(autos.ColumnNamesCSV())
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateCreateIfNotExists(object DatabaseMapped) (statementLabel, queryBody string, autos, writeCols *ColumnCollection) {
	cols := getCachedColumnCollectionFromInstance(object)

	writeCols = cols.WriteColumns()
	autos = cols.Autos()

	pks := cols.PrimaryKeys()
	tableName := TableName(object)

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range writeCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < writeCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(")")

	if pks.Len() > 0 {
		queryBodyBuffer.WriteString(" ON CONFLICT (")
		pkColumnNames := pks.ColumnNames()
		for i, name := range pkColumnNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkColumnNames)-1 {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
		queryBodyBuffer.WriteString(") DO NOTHING")
	}

	if autos.Len() > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(autos.ColumnNamesCSV())
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create_if_not_exists"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateCreateMany(objects interface{}) (queryBody string, writeCols *ColumnCollection, sliceValue reflect.Value) {
	sliceValue = reflectValue(objects)
	sliceType := reflectSliceType(objects)
	tableName := TableNameByType(sliceType)

	cols := getCachedColumnCollectionFromType(tableName, sliceType)
	writeCols = cols.WriteColumns()

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range writeCols.ColumnNames() {
		queryBodyBuffer.WriteString(name)
		if i < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}

	queryBodyBuffer.WriteString(") VALUES ")

	metaIndex := 1
	for x := 0; x < sliceValue.Len(); x++ {
		queryBodyBuffer.WriteString("(")
		for y := 0; y < writeCols.Len(); y++ {
			queryBodyBuffer.WriteString(fmt.Sprintf("$%d", metaIndex))
			metaIndex = metaIndex + 1
			if y < writeCols.Len()-1 {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
		queryBodyBuffer.WriteString(")")
		if x < sliceValue.Len()-1 {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}

	queryBody = queryBodyBuffer.String()
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateUpdate(object DatabaseMapped) (statementLabel, queryBody string, pks, writeCols *ColumnCollection) {
	tableName := TableName(object)

	cols := getCachedColumnCollectionFromInstance(object)

	pks = cols.PrimaryKeys()
	writeCols = cols.WriteColumns()

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("UPDATE ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" SET ")

	var writeColIndex int
	var col Column
	for ; writeColIndex < writeCols.Len(); writeColIndex++ {
		col = writeCols.columns[writeColIndex]
		queryBodyBuffer.WriteString(col.ColumnName)
		queryBodyBuffer.WriteString(" = $" + strconv.Itoa(writeColIndex+1))
		if writeColIndex != (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}

	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+(writeColIndex+1)))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_update"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateUpsert(object DatabaseMapped) (statementLabel, queryBody string, autos, writeCols *ColumnCollection) {
	tableName := TableName(object)
	cols := getCachedColumnCollectionFromInstance(object)
	updates := cols.NotReadOnly().NotAutos().NotPrimaryKeys().NotUniqueKeys()
	updateCols := updates.Columns()

	writeCols = cols.NotReadOnly().NotAutos()
	writeColNames := writeCols.ColumnNames()

	autos = cols.Autos()
	pks := cols.PrimaryKeys()
	pkNames := pks.ColumnNames()

	queryBodyBuffer := i.conn.bufferPool.Get()

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range writeColNames {
		queryBodyBuffer.WriteString(name)
		if i < len(writeColNames)-1 {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")

	for x := 0; x < writeCols.Len(); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (writeCols.Len() - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}

	queryBodyBuffer.WriteString(")")

	if pks.Len() > 0 {
		tokenMap := map[string]string{}
		for i, col := range writeCols.Columns() {
			tokenMap[col.ColumnName] = "$" + strconv.Itoa(i+1)
		}

		queryBodyBuffer.WriteString(" ON CONFLICT (")

		for i, name := range pkNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkNames)-1 {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
		queryBodyBuffer.WriteString(") DO UPDATE SET ")

		for i, col := range updateCols {
			queryBodyBuffer.WriteString(col.ColumnName + " = " + tokenMap[col.ColumnName])
			if i < (len(updateCols) - 1) {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
	}
	if autos.Len() > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(autos.ColumnNamesCSV())
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_upsert"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateExists(object DatabaseMapped) (statementLabel, queryBody string, pks *ColumnCollection, err error) {
	tableName := TableName(object)
	pks = getCachedColumnCollectionFromInstance(object).PrimaryKeys()
	if pks.Len() == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}
	queryBodyBuffer := i.conn.bufferPool.Get()
	queryBodyBuffer.WriteString("SELECT 1 FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_exists"
	queryBody = queryBodyBuffer.String()
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateDelete(object DatabaseMapped) (statementLabel, queryBody string, pks *ColumnCollection, err error) {
	tableName := TableName(object)
	pks = getCachedColumnCollectionFromInstance(object).PrimaryKeys()
	if len(pks.Columns()) == 0 {
		err = Error(ErrNoPrimaryKey)
		return
	}
	queryBodyBuffer := i.conn.bufferPool.Get()
	queryBodyBuffer.WriteString("DELETE FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks.Columns() {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (pks.Len() - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_delete"
	queryBody = queryBodyBuffer.String()
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

func (i *Invocation) generateTruncate(object DatabaseMapped) (statmentLabel, queryBody string) {
	tableName := TableName(object)

	queryBodyBuffer := i.conn.bufferPool.Get()
	queryBodyBuffer.WriteString("TRUNCATE ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	statmentLabel = tableName + "_truncate"
	i.conn.bufferPool.Put(queryBodyBuffer)
	return
}

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

func (i *Invocation) autoValues(autos *ColumnCollection) []interface{} {
	autoValues := make([]interface{}, autos.Len())
	for i, autoCol := range autos.Columns() {
		autoValues[i] = reflect.New(reflect.PtrTo(autoCol.FieldType)).Interface()
	}
	return autoValues
}

func (i *Invocation) setAutos(object DatabaseMapped, autos *ColumnCollection, autoValues []interface{}) (err error) {
	for index := 0; index < len(autoValues); index++ {
		err = autos.Columns()[index].SetValue(object, autoValues[index])
		if err != nil {
			err = Error(err)
			return
		}
	}
	return
}

func (i *Invocation) closeStatement(stmt *sql.Stmt, err error) error {
	// if we're within a transaction, DO NOT CLOSE THE STATEMENT.
	if stmt == nil || i.tx != nil {
		return err
	}
	// if the statement is cached, DO NOT CLOSE THE STATEMENT.
	if i.conn.planCache != nil && i.conn.planCache.Enabled() && i.cachedPlanKey != "" {
		return err
	}
	// close the statement.
	return exception.Nest(err, Error(stmt.Close()))
}

func (i *Invocation) start(statement string) (string, error) {
	if i.statementInterceptor != nil {
		var err error
		statement, err = i.statementInterceptor(i.cachedPlanKey, statement)
		if err != nil {
			return "", err
		}
	}
	if i.tracer != nil {
		i.traceFinisher = i.tracer.Query(i.context, i.conn, i, statement)
	}
	return statement, nil
}

func (i *Invocation) finish(statement string, r interface{}, err error) error {
	if i.cancel != nil {
		i.cancel()
	}
	if r != nil {
		err = exception.Nest(err, exception.New(r))
	}
	if i.conn.log != nil {
		i.conn.log.Trigger(
			logger.NewQueryEvent(statement, time.Now().UTC().Sub(i.startTime)).
				WithUsername(i.conn.config.GetUsername()).
				WithDatabase(i.conn.config.GetDatabase()).
				WithQueryLabel(i.cachedPlanKey).
				WithEngine(i.conn.config.GetEngine()).
				WithErr(err),
		)
	}
	if i.traceFinisher != nil {
		i.traceFinisher.Finish(err)
	}
	if err != nil {
		err = Error(err)
	}
	return err
}
