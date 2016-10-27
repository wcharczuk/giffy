package spiffy

import (
	"database/sql"
	"sync"
)

// NewStatementCache returns a new `StatementCache`.
func NewStatementCache(dbc *sql.DB) *StatementCache {
	return &StatementCache{
		dbc:       dbc,
		cacheLock: &sync.RWMutex{},
		cache:     make(map[string]*sql.Stmt),
	}
}

// StatementCache is a cache of prepared statements.
type StatementCache struct {
	dbc       *sql.DB
	cacheLock *sync.RWMutex
	cache     map[string]*sql.Stmt
}

// Close implements io.Closer.
func (sc *StatementCache) Close() error {
	return sc.Clear()
}

func (sc *StatementCache) closeAll() error {
	var err error
	for _, stmt := range sc.cache {
		err = stmt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Clear deletes all cached statements.
func (sc *StatementCache) Clear() error {
	sc.cacheLock.Lock()
	defer sc.cacheLock.Unlock()

	err := sc.closeAll()
	sc.cache = make(map[string]*sql.Stmt)
	return err
}

// HasStatement returns if the cache contains a statement.
func (sc *StatementCache) HasStatement(statement string) bool {
	return sc.getCachedStatement(statement) != nil
}

func (sc *StatementCache) getCachedStatement(statement string) *sql.Stmt {
	sc.cacheLock.RLock()
	defer sc.cacheLock.RUnlock()

	if stmt, hasStmt := sc.cache[statement]; hasStmt {
		return stmt
	}
	return nil
}

// Prepare returns a cached expression for a statement, or creates and caches a new one.
func (sc *StatementCache) Prepare(statement string) (*sql.Stmt, error) {
	cached := sc.getCachedStatement(statement)
	if cached != nil {
		return cached, nil
	}

	sc.cacheLock.Lock()
	defer sc.cacheLock.Unlock()

	// getCachedStatement without locking ...
	if stmt, hasStmt := sc.cache[statement]; hasStmt {
		return stmt, nil
	}

	stmt, err := sc.dbc.Prepare(statement)
	if err != nil {
		return nil, err
	}

	sc.cache[statement] = stmt
	return stmt, nil
}
