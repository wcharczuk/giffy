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

const (
	connectionErrorMessage = "invocation context; db connection is nil"
)

// Invocation is a specific operation against a context.
type Invocation struct {
	conn           *Connection
	ctx            context.Context
	tx             *sql.Tx
	fireEvents     bool
	statementLabel string
}

// WithCtx sets the ctx and returns a reference to the invocation.
func (i *Invocation) WithCtx(ctx context.Context) *Invocation {
	i.ctx = ctx
	return i
}

// Ctx returns the underlying context.
func (i *Invocation) Ctx() context.Context {
	return i.ctx
}

// FireEvents returns if events are enabled.
func (i *Invocation) FireEvents() bool {
	return i.fireEvents
}

// WithFireEvents sets the fire events property and returns an invocation.
func (i *Invocation) WithFireEvents(flag bool) *Invocation {
	i.fireEvents = flag
	return i
}

// WithLabel instructs the query generator to get or create a cached prepared statement.
func (i *Invocation) WithLabel(label string) *Invocation {
	i.statementLabel = label
	return i
}

// Label returns the statement / plan cache label for the context.
func (i *Invocation) Label() string {
	return i.statementLabel
}

// Tx returns the underlying transaction.
func (i *Invocation) Tx() *sql.Tx {
	return i.tx
}

// Prepare returns a cached or newly prepared statment plan for a given sql statement.
func (i *Invocation) Prepare(statement string) (*sql.Stmt, error) {
	if len(i.statementLabel) > 0 {
		return i.conn.PrepareCached(i.statementLabel, statement, i.tx)
	}
	return i.conn.Prepare(statement, i.tx)
}

// Exec executes a sql statement with a given set of arguments.
func (i *Invocation) Exec(statement string, args ...interface{}) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, statement, start) }()

	stmt, stmtErr := i.Prepare(statement)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}

	defer i.closeStatement(err, stmt)

	if _, execErr := stmt.Exec(args...); execErr != nil {
		err = exception.Wrap(execErr)
		if err != nil {
			i.invalidateCachedStatement()
		}
		return
	}

	return
}

// Query returns a new query object for a given sql query and arguments.
func (i *Invocation) Query(query string, args ...interface{}) *Query {
	return &Query{statement: query, args: args, start: time.Now(), conn: i.conn, ctx: i.ctx, tx: i.tx, fireEvents: i.fireEvents, statementLabel: i.statementLabel}
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (i *Invocation) Get(object DatabaseMapped, ids ...interface{}) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	var queryBody string
	start := time.Now()

	meta := getCachedColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_get", tableName)
	}

	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	columnNames := standardCols.ColumnNames()
	pks := standardCols.PrimaryKeys()
	if pks.Len() == 0 {
		err = exception.New("no primary key on object to get by.")
		return
	}

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range columnNames {
		queryBodyBuffer.WriteString(name)
		if i < (len(columnNames) - 1) {
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

	queryBody = queryBodyBuffer.String()
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer i.closeStatement(err, stmt)

	var rows *sql.Rows
	var queryErr error
	if i.ctx != nil {
		rows, queryErr = stmt.QueryContext(i.ctx, ids...)
	} else {
		rows, queryErr = stmt.Query(ids...)
	}

	if queryErr != nil {
		err = exception.Wrap(queryErr)
		i.invalidateCachedStatement()
		return
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.Nest(err, closeErr)
		}
	}()

	var popErr error
	if rows.Next() {
		if isPopulatable(object) {
			popErr = asPopulatable(object).Populate(rows)
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

// GetAll returns all rows of an object mapped table wrapped in a transaction.
func (i *Invocation) GetAll(collection interface{}) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName := TableNameByType(t)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_get_all", tableName)
	}

	meta := getCachedColumnCollectionFromType(tableName, t).NotReadOnly()

	columnNames := meta.ColumnNames()

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range columnNames {
		queryBodyBuffer.WriteString(name)
		if i < (len(columnNames) - 1) {
			queryBodyBuffer.WriteRune(runeComma)
		}
	}
	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		i.invalidateCachedStatement()
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var rows *sql.Rows
	var queryErr error
	if i.ctx != nil {
		rows, queryErr = stmt.QueryContext(i.ctx)
	} else {
		rows, queryErr = stmt.Query()
	}
	if queryErr != nil {
		err = exception.Wrap(queryErr)
		return
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.Nest(err, closeErr)
		}
	}()

	v, err := makeNewDatabaseMapped(t)
	if err != nil {
		return
	}
	isPopulatable := isPopulatable(v)

	var popErr error
	for rows.Next() {
		newObj, _ := makeNewDatabaseMapped(t)

		if isPopulatable {
			popErr = asPopulatable(newObj).Populate(rows)
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

// Create writes an object to the database within a transaction.
func (i *Invocation) Create(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotAutos()

	autos := cols.Autos()
	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_create", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range colNames {
		queryBodyBuffer.WriteString(name)
		if i < len(colNames)-1 {
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
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var execErr error
	if autos.Len() == 0 {
		if i.ctx != nil {
			_, execErr = stmt.ExecContext(i.ctx, colValues...)
		} else {
			_, execErr = stmt.Exec(colValues...)
		}

		if execErr != nil {
			err = exception.Wrap(execErr)
			i.invalidateCachedStatement()
			return
		}
	} else {
		autoValues := make([]interface{}, autos.Len())
		for i, autoCol := range autos.Columns() {
			autoValues[i] = reflect.New(reflect.PtrTo(autoCol.FieldType)).Interface()
		}

		if i.ctx != nil {
			execErr = stmt.QueryRowContext(i.ctx, colValues...).Scan(autoValues...)
		} else {
			execErr = stmt.QueryRow(colValues...).Scan(autoValues...)
		}

		if execErr != nil {
			err = exception.Wrap(execErr)
			return
		}

		for index := 0; index < len(autoValues); index++ {
			setErr := autos.Columns()[index].SetValue(object, autoValues[index])
			if setErr != nil {
				err = exception.Wrap(setErr)
				return
			}
		}
	}

	return nil
}

// CreateIfNotExists writes an object to the database if it does not already exist within a transaction.
func (i *Invocation) CreateIfNotExists(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotAutos()

	//NOTE: we're only using one.
	autos := cols.Autos()
	pks := cols.PrimaryKeys()
	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_create_if_not_exists", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range colNames {
		queryBodyBuffer.WriteString(name)
		if i < len(colNames)-1 {
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
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var execErr error
	if autos.Len() == 0 {
		if i.ctx != nil {
			_, execErr = stmt.ExecContext(i.ctx, colValues...)
		} else {
			_, execErr = stmt.Exec(colValues...)
		}
		if execErr != nil {
			err = exception.Wrap(execErr)
			i.invalidateCachedStatement()
			return
		}
	} else {
		autoValues := make([]interface{}, autos.Len())
		for i, autoCol := range autos.Columns() {
			autoValues[i] = reflect.New(reflect.PtrTo(autoCol.FieldType)).Interface()
		}

		if i.ctx != nil {
			execErr = stmt.QueryRowContext(i.ctx, colValues...).Scan(autoValues...)
		} else {
			execErr = stmt.QueryRow(colValues...).Scan(autoValues...)
		}

		if execErr != nil {
			err = exception.Wrap(execErr)
			return
		}

		for index := 0; index < len(autoValues); index++ {
			setErr := autos.Columns()[index].SetValue(object, autoValues[index])
			if setErr != nil {
				err = exception.Wrap(setErr)
				return
			}
		}
	}

	return nil
}

// CreateMany writes many an objects to the database within a transaction.
func (i *Invocation) CreateMany(objects interface{}) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	sliceValue := reflectValue(objects)
	if sliceValue.Len() == 0 {
		return nil
	}

	sliceType := reflectSliceType(objects)
	tableName := TableNameByType(sliceType)

	cols := getCachedColumnCollectionFromType(tableName, sliceType)
	writeCols := cols.NotReadOnly().NotAutos()

	//NOTE: we're only using one.
	//serials := cols.Serials()
	colNames := writeCols.ColumnNames()

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range colNames {
		queryBodyBuffer.WriteString(name)
		if i < len(colNames)-1 {
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
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var colValues []interface{}
	for row := 0; row < sliceValue.Len(); row++ {
		colValues = append(colValues, writeCols.ColumnValues(sliceValue.Index(row).Interface())...)
	}

	_, execErr := stmt.Exec(colValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
		return
	}

	return nil
}

// Update updates an object wrapped in a transaction.
func (i *Invocation) Update(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	tableName := TableName(object)
	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_update", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.WriteColumns()
	pks := cols.PrimaryKeys()
	updateCols := cols.UpdateColumns()
	updateValues := updateCols.ColumnValues(object)
	numColumns := writeCols.Len()

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("UPDATE ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" SET ")

	var writeColIndex int
	var col Column
	for ; writeColIndex < writeCols.Len(); writeColIndex++ {
		col = writeCols.columns[writeColIndex]
		queryBodyBuffer.WriteString(col.ColumnName)
		queryBodyBuffer.WriteString(" = $" + strconv.Itoa(writeColIndex+1))
		if writeColIndex != numColumns-1 {
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
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}

	defer func() { err = i.closeStatement(err, stmt) }()

	var execErr error
	if i.ctx != nil {
		_, execErr = stmt.ExecContext(i.ctx, updateValues...)
	} else {
		_, execErr = stmt.Exec(updateValues...)
	}
	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (i *Invocation) Exists(object DatabaseMapped) (exists bool, err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	tableName := TableName(object)
	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_exists", tableName)
	}
	cols := getCachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if pks.Len() == 0 {
		exists = false
		err = exception.New("No primary key on object.")
		return
	}

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

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

	queryBody = queryBodyBuffer.String()
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		exists = false
		err = exception.Wrap(stmtErr)
		return
	}

	defer func() { err = i.closeStatement(err, stmt) }()

	pkValues := pks.ColumnValues(object)
	var rows *sql.Rows
	var queryErr error
	if i.ctx != nil {
		rows, queryErr = stmt.QueryContext(i.ctx, pkValues...)
	} else {
		rows, queryErr = stmt.Query(pkValues...)
	}
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			err = exception.Nest(err, closeErr)
		}
	}()

	if queryErr != nil {
		exists = false
		err = exception.Wrap(queryErr)
		i.invalidateCachedStatement()
		return
	}

	exists = rows.Next()
	return
}

// Delete deletes an object from the database wrapped in a transaction.
func (i *Invocation) Delete(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_delete", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

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

	queryBody = queryBodyBuffer.String()
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	pkValues := pks.ColumnValues(object)

	var execErr error
	if i.ctx != nil {
		_, execErr = stmt.ExecContext(i.ctx, pkValues...)
	} else {
		_, execErr = stmt.Exec(pkValues...)
	}
	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
	}
	return
}

// Truncate completely empties a table in a single command.
func (i *Invocation) Truncate(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_truncate", tableName)
	}

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("TRUNCATE ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var execErr error
	if i.ctx != nil {
		_, execErr = stmt.ExecContext(i.ctx)
	} else {
		_, execErr = stmt.Exec()
	}

	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (i *Invocation) Upsert(object DatabaseMapped) (err error) {
	err = i.Validate()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.finalizer(recover(), err, logger.Query, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotAutos()

	conflictUpdateCols := cols.NotReadOnly().NotAutos().NotPrimaryKeys()

	serials := cols.Autos()
	pks := cols.PrimaryKeys()
	tableName := TableName(object)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_upsert", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.conn.bufferPool.Get()
	defer i.conn.bufferPool.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range colNames {
		queryBodyBuffer.WriteString(name)
		if i < len(colNames)-1 {
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
		pkColumnNames := pks.ColumnNames()
		for i, name := range pkColumnNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkColumnNames)-1 {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
		queryBodyBuffer.WriteString(") DO UPDATE SET ")

		conflictCols := conflictUpdateCols.Columns()
		for i, col := range conflictCols {
			queryBodyBuffer.WriteString(col.ColumnName + " = " + tokenMap[col.ColumnName])
			if i < (len(conflictCols) - 1) {
				queryBodyBuffer.WriteRune(runeComma)
			}
		}
	}

	var serial = serials.FirstOrDefault()
	if serials.Len() != 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(serial.ColumnName)
	}

	queryBody = queryBodyBuffer.String()

	stmt, stmtErr := i.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = i.closeStatement(err, stmt) }()

	var execErr error
	if serials.Len() != 0 {
		var id interface{}
		if i.ctx != nil {
			execErr = stmt.QueryRowContext(i.ctx, colValues...).Scan(&id)
		} else {
			execErr = stmt.QueryRow(colValues...).Scan(&id)
		}
		if execErr != nil {
			err = exception.Wrap(execErr)
			i.invalidateCachedStatement()
			return
		}
		setErr := serial.SetValue(object, id)
		if setErr != nil {
			err = exception.Wrap(setErr)
			return
		}
	} else {
		if i.ctx != nil {
			_, execErr = stmt.ExecContext(i.ctx, colValues...)
		} else {
			_, execErr = stmt.Exec(colValues...)
		}
		if execErr != nil {
			err = exception.Wrap(execErr)
			return
		}
	}

	return nil
}

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

// Validate validates the invocation is ready
func (i *Invocation) Validate() error {
	if i.conn == nil {
		return exception.Newf(connectionErrorMessage)
	}
	return nil
}

func (i *Invocation) invalidateCachedStatement() {
	if i.conn.useStatementCache && len(i.statementLabel) > 0 {
		i.conn.statementCache.InvalidateStatement(i.statementLabel)
	}
}

func (i *Invocation) closeStatement(err error, stmt *sql.Stmt) error {
	if !i.conn.useStatementCache {
		closeErr := stmt.Close()
		if closeErr != nil {
			return exception.Nest(err, closeErr)
		}
	}
	return err
}

func (i *Invocation) finalizer(r interface{}, err error, flag logger.Flag, statement string, start time.Time) error {
	if r != nil {
		recoveryException := exception.New(r)
		err = exception.Nest(err, recoveryException)
	}
	if i.fireEvents {
		i.conn.fireEvent(flag, statement, time.Now().Sub(start), err, i.statementLabel)
	}
	i.statementLabel = ""
	return err
}
