package spiffy

import "sync"

var (
	defaultAlias     string
	defaultAliasLock = sync.Mutex{}
	dbAliasesLock    = sync.Mutex{}
	dbAliases        = make(map[string]*DbConnection)
)

// CreateDbAlias allows you to set up a connection for later use via an alias.
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//
// You can later set an alias as 'default' and refer to it using `spiffy.DefaultDb()`.
func CreateDbAlias(alias string, prototype *DbConnection) {
	dbAliasesLock.Lock()
	defer dbAliasesLock.Unlock()
	dbAliases[alias] = prototype
}

// Alias fetches a connection by its alias.
//
//	spiffy.Alias("logging").Create(&object)
//
// Alternately, if you've set the alias as 'default' you can just refer to it via. `DefaultDb()`
func Alias(alias string) *DbConnection {
	return dbAliases[alias]
}

// SetDefaultAlias sets an alias created with `CreateDbAlias` as default. This lets you refer to it later via. `DefaultDb()`
//
//	spiffy.CreateDbAlias("main", spiffy.NewDbConnection("localhost", "test_db", "", ""))
//	spiffy.SetDefaultAlias("main")
//	execErr := spiffy.DefaultDb().Execute("select 'ok!')
//
// This will then let you refer to the alias via. `DefaultDb()`
func SetDefaultAlias(alias string) {
	defaultAliasLock.Lock()
	defer defaultAliasLock.Unlock()
	defaultAlias = alias
}

// SetDefaultDb sets the default db to a connection
func SetDefaultDb(conn *DbConnection) error {
	aliasName := UUIDv4().ToShortString()
	CreateDbAlias(aliasName, conn)
	SetDefaultAlias(aliasName)
	_, err := conn.Open()
	return err
}

// DefaultDb returns a reference to the DbConnection set as default.
//
//	spiffy.DefaultDb().Exec("select 'ok!")
//
// Note: you must set up the default with `SetDefaultAlias()` before using DefaultDb.
func DefaultDb() *DbConnection {
	if len(defaultAlias) != 0 {
		return dbAliases[defaultAlias]
	}
	return nil
}
