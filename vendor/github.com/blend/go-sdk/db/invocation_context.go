package db

import (
	"context"
	"database/sql"
)

// NewInvocationContext returns a new invocation context.
func NewInvocationContext(conn *Connection) *InvocationContext {
	return &InvocationContext{
		conn:       conn,
		fireEvents: true,
	}
}

// InvocationContext represents an invocation context.
// It rolls both the underlying connection and an optional tx into one struct.
// The motivation here is so that if you have datamanager functions they can be
// used across databases, and don't assume internally which db they talk to.
type InvocationContext struct {
	conn       *Connection
	ctx        context.Context
	tx         *sql.Tx
	fireEvents bool
}

// WithCtx sets the db context.
func (ic *InvocationContext) WithCtx(ctx context.Context) *InvocationContext {
	ic.ctx = ctx
	return ic
}

// Ctx returns the context on the invocation context.
func (ic *InvocationContext) Ctx() context.Context {
	return ic.ctx
}

// FireEvents returns if events are enabled.
func (ic *InvocationContext) FireEvents() bool {
	return ic.fireEvents
}

// WithFireEvents sets the `FireEvents` property.
func (ic *InvocationContext) WithFireEvents(flag bool) *InvocationContext {
	ic.fireEvents = flag
	return ic
}

// WithConnection sets the connection for the context.
func (ic *InvocationContext) WithConnection(conn *Connection) *InvocationContext {
	ic.conn = conn
	return ic
}

// Connection returns the underlying connection for the context.
func (ic *InvocationContext) Connection() *Connection {
	return ic.conn
}

// InTx isolates a context to a transaction.
func (ic *InvocationContext) InTx(txs ...*sql.Tx) *InvocationContext {
	if len(txs) > 0 {
		ic.tx = txs[0]
		return ic
	}
	return ic
}

// Tx returns the transction for the context.
func (ic *InvocationContext) Tx() *sql.Tx {
	return ic.tx
}

// Commit calls `Commit()` on the underlying transaction.
func (ic *InvocationContext) Commit() error {
	if ic.tx == nil {
		return nil
	}
	return ic.tx.Commit()
}

// Rollback calls `Rollback()` on the underlying transaction.
func (ic *InvocationContext) Rollback() error {
	if ic.tx == nil {
		return nil
	}
	return ic.tx.Rollback()
}

// Invoke starts a new invocation.
func (ic *InvocationContext) Invoke() *Invocation {
	return &Invocation{conn: ic.conn, ctx: ic.ctx, tx: ic.tx, fireEvents: ic.fireEvents}
}
