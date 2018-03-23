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
	State() State
}

// Tx returns the transaction for the request.
// keys is an optional parameter used for additional arbitrary transactions
func Tx(sp StateProvider, keys ...string) *sql.Tx {
	if sp == nil {
		return nil
	}
	return TxFromState(sp.State(), keys...)
}

// TxFromState returns a tx from a state bag.
func TxFromState(state State, keys ...string) *sql.Tx {
	if state == nil {
		return nil
	}

	key := StateKeyTx
	if keys != nil && len(keys) > 0 {
		key = StateKeyPrefixTx + keys[0]
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

	key := "tx"
	if keys != nil && len(keys) > 0 {
		key = StateKeyPrefixTx + keys[0]
	}
	state := sp.State()
	if state == nil {
		return nil
	}
	state[key] = tx
	return sp
}
