package web

import "database/sql"

const (
	// StateKeyTx is the app state key for a transaction.
	StateKeyTx = "tx"
)

// State is the collection of state objects on a context.
type State map[string]interface{}

// StateProvider provide states, an example is Ctx
type StateProvider interface {
	State() State
}

// StateValueProvider is a type that provides a state value.
type StateValueProvider interface {
	StateValue(key string) interface{}
}

// Tx returns the transaction for the request.
// keys is an optional parameter used for additional arbitrary transactions
func Tx(sp StateProvider, optionalKey ...string) *sql.Tx {
	if sp == nil {
		return nil
	}
	return TxFromState(sp.State(), optionalKey...)
}

// TxFromState returns a tx from a state bag.
func TxFromState(state State, keys ...string) *sql.Tx {
	if state == nil {
		return nil
	}

	key := StateKeyTx
	if len(keys) > 0 {
		key = keys[0]
	}
	if typed, isTyped := state[key].(*sql.Tx); isTyped {
		return typed
	}

	return nil
}

// WithTx sets a transaction on the state provider.
func WithTx(sp StateProvider, tx *sql.Tx, keys ...string) StateProvider {
	if sp == nil {
		return nil
	}
	state := sp.State()
	if state == nil {
		return nil
	}
	WithTxForState(state, tx, keys...)
	return sp
}

// WithTxForState injects a tx into a statebag.
func WithTxForState(state State, tx *sql.Tx, keys ...string) {
	key := StateKeyTx
	if len(keys) > 0 {
		key = keys[0]
	}
	state[key] = tx
}
