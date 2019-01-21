package db

import (
	"sync"
)

var (
	defaultConnection *Connection
	defaultLock       = sync.Mutex{}
)

// SetDefault sets an alias created with `CreateDbAlias` as default. This lets you refer to it later via. `Default()`
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//	spiffy.SetDefault("main")
//	execErr := spiffy.Default().Execute("select 'ok!')
//
// This will then let you refer to the alias via. `Default()`
func SetDefault(conn *Connection) {
	defaultLock.Lock()
	defaultConnection = conn
	defaultLock.Unlock()
}

// Default returns a reference to the DbConnection set as default.
//
//	spiffy.Default().Exec("select 'ok!")
//
func Default() *Connection {
	return defaultConnection
}

// OpenDefault sets the default connection and opens it.
func OpenDefault(conn *Connection) error {
	err := conn.Open()
	if err != nil {
		return err
	}
	SetDefault(conn)
	return nil
}
