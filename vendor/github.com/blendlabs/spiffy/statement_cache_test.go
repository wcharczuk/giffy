package spiffy

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestStatementCachePrepare(t *testing.T) {
	assert := assert.New(t)

	sc := newStatementCache(Default().Connection)
	query := "select 'ok'"
	stmt, err := sc.Prepare(query, query)

	assert.Nil(err)
	assert.NotNil(stmt)
	assert.True(sc.HasStatement(query))

	// shoul result in cache hit
	stmt, err = sc.Prepare(query, query)
	assert.NotNil(stmt)
	assert.True(sc.HasStatement(query))
}
