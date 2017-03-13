package spiffy

import (
	"database/sql"
	"sync"
)

// newStatementCache returns a new `StatementCache`.
func newStatementCache(dbc *sql.DB) *StatementCache {
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
	err := sc.closeAll()
	sc.cache = make(map[string]*sql.Stmt)
	sc.cacheLock.Unlock()
	return err
}

// HasStatement returns if the cache contains a statement.
func (sc *StatementCache) HasStatement(statementID string) bool {
	return sc.getCachedStatement(statementID) != nil
}

// InvalidateStatement removes a statement from the cache.
func (sc *StatementCache) InvalidateStatement(statementID string) {
	sc.cacheLock.Lock()
	if _, hasStatement := sc.cache[statementID]; hasStatement {
		delete(sc.cache, statementID)
	}
	sc.cacheLock.Unlock()
}

func (sc *StatementCache) getCachedStatement(statementID string) *sql.Stmt {
	sc.cacheLock.RLock()

	if stmt, hasStmt := sc.cache[statementID]; hasStmt {
		sc.cacheLock.RUnlock()
		return stmt
	}
	sc.cacheLock.RUnlock()
	return nil
}

// Prepare returns a cached expression for a statement, or creates and caches a new one.
func (sc *StatementCache) Prepare(id, statementProvider string) (*sql.Stmt, error) {
	cached := sc.getCachedStatement(id)
	if cached != nil {
		return cached, nil
	}

	sc.cacheLock.Lock()
	defer sc.cacheLock.Unlock()
	if stmt, hasStmt := sc.cache[id]; hasStmt {
		return stmt, nil
	}

	stmt, err := sc.dbc.Prepare(statementProvider)
	if err != nil {
		return nil, err
	}

	sc.cache[id] = stmt
	return stmt, nil
}
