package spiffy

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

const (
	connectionErrorMessage = "invocation context; db connection is nil"
)

// Invocation is a specific operation against a context.
type Invocation struct {
	db             *DB
	fireEvents     bool
	statementLabel string
	err            error
}

// Err returns the context's error.
func (i *Invocation) Err() error {
	return i.err
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
	return i.db.tx
}

// Prepare returns a cached or newly prepared statment plan for a given sql statement.
func (i *Invocation) Prepare(statement string) (*sql.Stmt, error) {
	if i.err != nil {
		return nil, i.err
	}
	if len(i.statementLabel) > 0 {
		return i.db.conn.PrepareCached(i.statementLabel, statement, i.db.tx)
	}
	return i.db.conn.Prepare(statement, i.db.tx)
}

// Exec executes a sql statement with a given set of arguments.
func (i *Invocation) Exec(statement string, args ...interface{}) (err error) {
	err = i.check()
	if err != nil {
		return
	}

	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, statement, start) }()

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
	return &Query{statement: query, args: args, start: time.Now(), db: i.db, err: i.check(), statementLabel: i.statementLabel}
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (i *Invocation) Get(object DatabaseMapped, ids ...interface{}) (err error) {
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := getCachedColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	tableName := object.TableName()

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_get", tableName)
	}

	columnNames := standardCols.ColumnNames()
	pks := standardCols.PrimaryKeys()
	if pks.Len() == 0 {
		err = exception.New("no primary key on object to get by.")
		return
	}

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	rows, queryErr := stmt.Query(ids...)
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
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_get_all", tableName)
	}

	meta := getCachedColumnCollectionFromType(tableName, t).NotReadOnly()

	columnNames := meta.ColumnNames()

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	rows, queryErr := stmt.Query()
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
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_create", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	if serials.Len() > 0 {
		serial := serials.FirstOrDefault()
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

	if serials.Len() == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			err = exception.Wrap(execErr)
			i.invalidateCachedStatement()
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

// CreateIfNotExists writes an object to the database if it does not already exist within a transaction.
func (i *Invocation) CreateIfNotExists(object DatabaseMapped) (err error) {
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	pks := cols.PrimaryKeys()
	tableName := object.TableName()

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_create_if_not_exists", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	if serials.Len() > 0 {
		serial := serials.FirstOrDefault()
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

	if serials.Len() == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			err = exception.Wrap(execErr)
			i.invalidateCachedStatement()
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

// CreateMany writes many an objects to the database within a transaction.
func (i *Invocation) CreateMany(objects interface{}) (err error) {
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	sliceValue := reflectValue(objects)
	if sliceValue.Len() == 0 {
		return nil
	}

	sliceType := reflectSliceType(objects)
	tableName, err := TableName(sliceType)
	if err != nil {
		return
	}

	cols := getCachedColumnCollectionFromType(tableName, sliceType)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	//serials := cols.Serials()
	colNames := writeCols.ColumnNames()

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	tableName := object.TableName()
	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_update", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.WriteColumns()
	pks := cols.PrimaryKeys()
	updateCols := cols.UpdateColumns()
	updateValues := updateCols.ColumnValues(object)
	numColumns := writeCols.Len()

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	_, execErr := stmt.Exec(updateValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (i *Invocation) Exists(object DatabaseMapped) (exists bool, err error) {
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	tableName := object.TableName()
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

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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
	rows, queryErr := stmt.Query(pkValues...)
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
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	tableName := object.TableName()

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_delete", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	_, execErr := stmt.Exec(pkValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		i.invalidateCachedStatement()
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (i *Invocation) Upsert(object DatabaseMapped) (err error) {
	err = i.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = i.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	conflictUpdateCols := cols.NotReadOnly().NotSerials().NotPrimaryKeys()

	serials := cols.Serials()
	pks := cols.PrimaryKeys()
	tableName := object.TableName()

	if len(i.statementLabel) == 0 {
		i.statementLabel = fmt.Sprintf("%s_upsert", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := i.db.conn.bufferPool.Get()
	defer i.db.conn.bufferPool.Put(queryBodyBuffer)

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

	if serials.Len() != 0 {
		var id interface{}
		execErr := stmt.QueryRow(colValues...).Scan(&id)
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
		_, execErr := stmt.Exec(colValues...)
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

func (i *Invocation) check() error {
	if i.db == nil {
		return exception.Newf(connectionErrorMessage)
	}
	if i.db.conn == nil {
		return exception.Newf(connectionErrorMessage)
	}
	if i.err != nil {
		return i.err
	}
	return nil
}

func (i *Invocation) invalidateCachedStatement() {
	if i.db.conn.useStatementCache && len(i.statementLabel) > 0 {
		i.db.conn.statementCache.InvalidateStatement(i.statementLabel)
	}
}

func (i *Invocation) closeStatement(err error, stmt *sql.Stmt) error {
	if !i.db.conn.useStatementCache {
		closeErr := stmt.Close()
		if closeErr != nil {
			return exception.Nest(err, closeErr)
		}
	}
	i.statementLabel = ""
	return err
}

func (i *Invocation) panicHandler(r interface{}, err error, eventFlag logger.EventFlag, statement string, start time.Time) error {
	if r != nil {
		recoveryException := exception.New(r)
		return exception.Nest(err, recoveryException)
	}
	if i.fireEvents {
		i.db.conn.fireEvent(eventFlag, statement, time.Now().Sub(start), err, i.statementLabel)
	}
	return err
}
