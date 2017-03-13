package spiffy

import "sync"

var (
	defaultConnection *Connection
	defaultLock       = sync.Mutex{}
	aliasesLock       = sync.Mutex{}
	aliases           = make(map[string]*Connection)
)

// CreateAlias allows you to set up a connection for later use via an alias.
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//
// You can later set an alias as 'default' and refer to it using `spiffy.DefaultDb()`.
func CreateAlias(alias string, prototype *Connection) {
	aliasesLock.Lock()
	aliases[alias] = prototype
	aliasesLock.Unlock()
}

// Alias fetches a connection by its alias.
//
//	spiffy.Alias("logging").Create(&object)
//
// Alternately, if you've set the alias as 'default' you can just refer to it via. `DefaultDb()`
func Alias(alias string) *Connection {
	return aliases[alias]
}

// SetDefault sets an alias created with `CreateDbAlias` as default. This lets you refer to it later via. `DefaultDb()`
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//	spiffy.SetDefault("main")
//	execErr := spiffy.DefaultDb().Execute("select 'ok!')
//
// This will then let you refer to the alias via. `DefaultDb()`
func SetDefault(conn *Connection) {
	defaultLock.Lock()
	defaultConnection = conn
	defaultLock.Unlock()
}

// InitDefault sets the default connection and opens it.
func InitDefault(conn *Connection) error {
	SetDefault(conn)
	_, err := conn.Open()
	return err
}

// Default returns a reference to the DbConnection set as default.
//
//	spiffy.DefaultDb().Exec("select 'ok!")
//
// Note: you must set up the default with `SetDefaultAlias()` before using DefaultDb.
func Default() *Connection {
	return defaultConnection
}

// DB is an alias to DefaultDb.
func DB() *Connection {
	return defaultConnection
}
