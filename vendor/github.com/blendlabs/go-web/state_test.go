package web

import (
	"database/sql"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestTx(t *testing.T) {
	assert := assert.New(t)

	testTx := &sql.Tx{}
	testAdditionalTx := &sql.Tx{}
	testCtx := NewCtx(nil, nil, nil, nil).WithTx(testTx).WithTx(testAdditionalTx, "additional")

	assert.NotNil(Tx(testCtx))
	assert.NotNil(Tx(testCtx, "additional"))
}

func TestWithTx(t *testing.T) {
	assert := assert.New(t)

	testTx := &sql.Tx{}
	testAdditionalTx := &sql.Tx{}
	testCtx := WithTx(NewCtx(nil, nil, nil, nil), testTx)
	testCtx = WithTx(testCtx, testAdditionalTx, "additional")

	assert.NotNil(Tx(testCtx))
	assert.NotNil(Tx(testCtx, "additional"))
}
