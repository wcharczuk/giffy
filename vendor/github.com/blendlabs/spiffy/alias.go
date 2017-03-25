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
//	spiffy.CreateAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//
func CreateAlias(alias string, prototype *Connection) {
	aliasesLock.Lock()
	aliases[alias] = prototype
	aliasesLock.Unlock()
}

// Alias fetches a connection by its alias.
//
//	spiffy.Alias("logging").Create(&object)
//
func Alias(alias string) *Connection {
	return aliases[alias]
}

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

// InitDefault sets the default connection and opens it.
func InitDefault(conn *Connection) error {
	SetDefault(conn)
	_, err := conn.Open()
	return err
}

// Default returns a reference to the DbConnection set as default.
//
//	spiffy.Default().Exec("select 'ok!")
//
func Default() *Connection {
	return defaultConnection
}

// DB is an alias to Default.
func DB() *Connection {
	return defaultConnection
}
