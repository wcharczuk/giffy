package web

import "database/sql"

const (
	// StateKeyPrefixTx is the prefix for keys for arbitrary db txs stored in states
	StateKeyPrefixTx = "tx-"
	// StateKeyTx is the app state key for a transaction.
	StateKeyTx = "tx"
)

// StateProvider provide states, an example is Ctx
type StateProvider interface {
	State(string) interface{}
	SetState(key string, value interface{})
}

// Tx returns the transaction for the request.
// keys is an optional parameter used for additional arbitrary transactions
func Tx(sp StateProvider, keys ...string) *sql.Tx {
	if sp == nil {
		return nil
	}
	key := StateKeyTx
	if keys != nil && len(keys) > 0 {
		key = StateKeyPrefixTx + keys[0]
	}
	if typed, isTyped := sp.State(key).(*sql.Tx); isTyped {
		return typed
	}

	return nil
}

// WithTx sets a transaction on the state provider.
func WithTx(sp StateProvider, tx *sql.Tx, keys ...string) StateProvider {
	key := "tx"
	if keys != nil && len(keys) > 0 {
		key = StateKeyPrefixTx + keys[0]
	}
	sp.SetState(key, tx)

	return sp
}
