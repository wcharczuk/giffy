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

// NewCtx returns a new ctx.
func NewCtx() *Ctx {
	return &Ctx{}
}

// Ctx represents a connection context.
type Ctx struct {
	conn           *Connection
	tx             *sql.Tx
	fireEvents     bool
	statementLabel string
	err            error
}

// WithConn sets the connection for the context.
func (c *Ctx) WithConn(conn *Connection) *Ctx {
	c.conn = conn
	return c
}

// Conn returns the underlying connection for the context.
func (c *Ctx) Conn() *Connection {
	return c.conn
}

// InTx isolates a context to a transaction.
// The order precedence of the three main transaction sources are as follows:
// - InTx(...) transaction arguments will be used above everything else
// - an existing transaction on the context (i.e. if you call `.InTx().InTx()`)
// - beginning a new transaction with the connection
func (c *Ctx) InTx(txs ...*sql.Tx) *Ctx {
	if len(txs) > 0 {
		c.tx = txs[0]
		return c
	}
	if c.tx != nil {
		return c
	}
	if c.conn == nil {
		c.err = exception.Newf(connectionErrorMessage)
		return c
	}
	c.tx, c.err = c.conn.Begin()
	return c
}

// Tx returns the transction for the context.
func (c *Ctx) Tx() *sql.Tx {
	return c.tx
}

// Err returns the context's error.
func (c *Ctx) Err() error {
	return c.err
}

// WithLabel instructs the query generator to get or create a cached prepared statement.
func (c *Ctx) WithLabel(label string) *Ctx {
	c.statementLabel = label
	return c
}

// Label returns the statement / plan cache label for the context.
func (c *Ctx) Label() string {
	return c.statementLabel
}

// Commit calls `Commit()` on the underlying transaction.
func (c *Ctx) Commit() error {
	if c.tx == nil {
		return nil
	}
	return c.tx.Commit()
}

// Rollback calls `Rollback()` on the underlying transaction.
func (c *Ctx) Rollback() error {
	if c.tx == nil {
		return nil
	}
	return c.tx.Rollback()
}

// Prepare returns a cached or newly prepared statment plan for a given sql statement.
func (c *Ctx) Prepare(statement string) (*sql.Stmt, error) {
	if len(c.statementLabel) > 0 {
		return c.conn.PrepareCached(c.statementLabel, statement, c.tx)
	}
	return c.conn.Prepare(statement, c.tx)
}

// Exec executes a sql statement with a given set of arguments.
func (c *Ctx) Exec(statement string, args ...interface{}) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, statement, start) }()

	stmt, stmtErr := c.Prepare(statement)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}

	defer c.closeStatement(err, stmt)

	if _, execErr := stmt.Exec(args...); execErr != nil {
		err = exception.Wrap(execErr)
		if err != nil {
			c.invalidateCachedStatement()
		}
		return
	}

	return
}

// Query returns a new query object for a given sql query and arguments.
func (c *Ctx) Query(query string, args ...interface{}) *Query {
	return &Query{statement: query, args: args, start: time.Now(), ctx: c, err: c.check()}
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (c *Ctx) Get(object DatabaseMapped, ids ...interface{}) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	if ids == nil {
		return exception.New("invalid `ids` parameter.")
	}

	meta := getCachedColumnCollectionFromInstance(object)
	standardCols := meta.NotReadOnly()
	tableName := object.TableName()

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_get", tableName)
	}

	columnNames := standardCols.ColumnNames()
	pks := standardCols.PrimaryKeys()
	if pks.Len() == 0 {
		err = exception.New("no primary key on object to get by.")
		return
	}

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer c.closeStatement(err, stmt)

	rows, queryErr := stmt.Query(ids...)
	if queryErr != nil {
		err = exception.Wrap(queryErr)
		c.invalidateCachedStatement()
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
func (c *Ctx) GetAll(collection interface{}) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	collectionValue := reflectValue(collection)
	t := reflectSliceType(collection)
	tableName, _ := TableName(t)

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_get_all", tableName)
	}

	meta := getCachedColumnCollectionFromType(tableName, t).NotReadOnly()

	columnNames := meta.ColumnNames()

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		c.invalidateCachedStatement()
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

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
func (c *Ctx) Create(object DatabaseMapped) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	tableName := object.TableName()

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_create", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

	if serials.Len() == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			err = exception.Wrap(execErr)
			c.invalidateCachedStatement()
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
func (c *Ctx) CreateIfNotExists(object DatabaseMapped) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	//NOTE: we're only using one.
	serials := cols.Serials()
	pks := cols.PrimaryKeys()
	tableName := object.TableName()

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_create_if_not_exists", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

	if serials.Len() == 0 {
		_, execErr := stmt.Exec(colValues...)
		if execErr != nil {
			err = exception.Wrap(execErr)
			c.invalidateCachedStatement()
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
func (c *Ctx) CreateMany(objects interface{}) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

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

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

	var colValues []interface{}
	for row := 0; row < sliceValue.Len(); row++ {
		colValues = append(colValues, writeCols.ColumnValues(sliceValue.Index(row).Interface())...)
	}

	_, execErr := stmt.Exec(colValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		c.invalidateCachedStatement()
		return
	}

	return nil
}

// Update updates an object wrapped in a transaction.
func (c *Ctx) Update(object DatabaseMapped) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	tableName := object.TableName()
	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_update", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.WriteColumns()
	pks := cols.PrimaryKeys()
	updateCols := cols.UpdateColumns()
	updateValues := updateCols.ColumnValues(object)
	numColumns := writeCols.Len()

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}

	defer func() { err = c.closeStatement(err, stmt) }()

	_, execErr := stmt.Exec(updateValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		c.invalidateCachedStatement()
		return
	}

	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (c *Ctx) Exists(object DatabaseMapped) (exists bool, err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagQuery, queryBody, start) }()

	tableName := object.TableName()
	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_exists", tableName)
	}
	cols := getCachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if pks.Len() == 0 {
		exists = false
		err = exception.New("No primary key on object.")
		return
	}

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		exists = false
		err = exception.Wrap(stmtErr)
		return
	}

	defer func() { err = c.closeStatement(err, stmt) }()

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
		c.invalidateCachedStatement()
		return
	}

	exists = rows.Next()
	return
}

// Delete deletes an object from the database wrapped in a transaction.
func (c *Ctx) Delete(object DatabaseMapped) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	tableName := object.TableName()

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_delete", tableName)
	}

	cols := getCachedColumnCollectionFromInstance(object)
	pks := cols.PrimaryKeys()

	if len(pks.Columns()) == 0 {
		err = exception.New("No primary key on object.")
		return
	}

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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
	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

	pkValues := pks.ColumnValues(object)

	_, execErr := stmt.Exec(pkValues...)
	if execErr != nil {
		err = exception.Wrap(execErr)
		c.invalidateCachedStatement()
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it wrapped in a transaction.
func (c *Ctx) Upsert(object DatabaseMapped) (err error) {
	err = c.check()
	if err != nil {
		return
	}

	var queryBody string
	start := time.Now()
	defer func() { err = c.panicHandler(recover(), err, EventFlagExecute, queryBody, start) }()

	cols := getCachedColumnCollectionFromInstance(object)
	writeCols := cols.NotReadOnly().NotSerials()

	conflictUpdateCols := cols.NotReadOnly().NotSerials().NotPrimaryKeys()

	serials := cols.Serials()
	pks := cols.PrimaryKeys()
	tableName := object.TableName()

	if len(c.statementLabel) == 0 {
		c.statementLabel = fmt.Sprintf("%s_upsert", tableName)
	}

	colNames := writeCols.ColumnNames()
	colValues := writeCols.ColumnValues(object)

	queryBodyBuffer := c.conn.bufferPool.Get()
	defer c.conn.bufferPool.Put(queryBodyBuffer)

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

	stmt, stmtErr := c.Prepare(queryBody)
	if stmtErr != nil {
		err = exception.Wrap(stmtErr)
		return
	}
	defer func() { err = c.closeStatement(err, stmt) }()

	if serials.Len() != 0 {
		var id interface{}
		execErr := stmt.QueryRow(colValues...).Scan(&id)
		if execErr != nil {
			err = exception.Wrap(execErr)
			c.invalidateCachedStatement()
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

func (c *Ctx) check() error {
	if c.conn == nil {
		return exception.Newf(connectionErrorMessage)
	}
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *Ctx) invalidateCachedStatement() {
	if c.conn.useStatementCache && len(c.statementLabel) > 0 {
		c.conn.statementCache.InvalidateStatement(c.statementLabel)
	}
}

func (c *Ctx) closeStatement(err error, stmt *sql.Stmt) error {
	if !c.conn.useStatementCache {
		closeErr := stmt.Close()
		if closeErr != nil {
			return exception.Nest(err, closeErr)
		}
	}
	c.statementLabel = ""
	return err
}

func (c *Ctx) panicHandler(r interface{}, err error, eventFlag logger.EventFlag, statement string, start time.Time) error {
	if r != nil {
		recoveryException := exception.New(r)
		return exception.Nest(err, recoveryException)
	}
	if c.fireEvents {
		c.conn.fireEvent(eventFlag, statement, time.Now().Sub(start), err, c.statementLabel)
	}
	return err
}
