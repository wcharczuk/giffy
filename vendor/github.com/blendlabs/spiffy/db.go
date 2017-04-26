package spiffy

import (
	"database/sql"

	exception "github.com/blendlabs/go-exception"
)

// NewDB returns a new DB.
func NewDB() *DB {
	return &DB{}
}

// DB represents a connection context.
// It rolls both the underlying connection and an optional tx into one struct.
// The motivation here is so that if you have datamanager functions they can be
// used across databases, and don't assume internally which db they talk to.
type DB struct {
	conn *Connection
	tx   *sql.Tx
	err  error
}

// WithConn sets the connection for the context.
func (db *DB) WithConn(conn *Connection) *DB {
	db.conn = conn
	return db
}

// Conn returns the underlying connection for the context.
func (db *DB) Conn() *Connection {
	return db.conn
}

// InTx isolates a context to a transaction.
// The order precedence of the three main transaction sources are as follows:
// - InTx(...) transaction arguments will be used above everything else
// - an existing transaction on the context (i.e. if you call `.InTx().InTx()`)
// - beginning a new transaction with the connection
func (db *DB) InTx(txs ...*sql.Tx) *DB {
	if len(txs) > 0 {
		db.tx = txs[0]
		return db
	}
	if db.tx != nil {
		return db
	}
	if db.conn == nil {
		db.err = exception.Newf(connectionErrorMessage)
		return db
	}
	db.tx, db.err = db.conn.Begin()
	return db
}

// Tx returns the transction for the context.
func (db *DB) Tx() *sql.Tx {
	return db.tx
}

// Commit calls `Commit()` on the underlying transaction.
func (db *DB) Commit() error {
	if db.tx == nil {
		return nil
	}
	return db.tx.Commit()
}

// Rollback calls `Rollback()` on the underlying transaction.
func (db *DB) Rollback() error {
	if db.tx == nil {
		return nil
	}
	return db.tx.Rollback()
}

// Err returns the carried error.
func (db *DB) Err() error {
	return db.err
}

// Invoke starts a new invocation.
func (db *DB) Invoke() *Invocation {
	return &Invocation{db: db, err: db.err}
}
