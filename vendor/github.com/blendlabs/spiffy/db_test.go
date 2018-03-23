package spiffy

import (
	"fmt"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestCtxInTxUsesArguments(t *testing.T) {
	assert := assert.New(t)
	tx, err := Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	withTx := NewDB().InTx(tx)
	assert.NotNil(withTx.tx)
}

func TestCtxInTxReturnsAnExistingTransaction(t *testing.T) {
	assert := assert.New(t)
	tx, err := Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	withTx := NewDB().InTx(tx).InTx()
	assert.NotNil(withTx.tx)
	assert.Equal(tx, withTx.Tx())
}

func TestCtxInTx(t *testing.T) {
	assert := assert.New(t)

	withTx := NewDB().WithConn(Default()).InTx()
	defer withTx.Rollback()
	assert.NotNil(withTx.tx)
}

func TestCtxInTxWithoutConnection(t *testing.T) {
	assert := assert.New(t)

	withTx := NewDB().InTx()
	assert.Nil(withTx.tx)
	assert.NotNil(withTx.err)
}

func TestCtxInvoke(t *testing.T) {
	assert := assert.New(t)

	inv := NewDB().WithConn(Default()).Invoke()
	assert.Nil(inv.check())
}

func TestCtxInvokeError(t *testing.T) {
	assert := assert.New(t)

	inv := NewDB().Invoke()
	assert.NotNil(inv.check(), "should fail the connection not nil check")
}

func TestCtxInvokeCarriesError(t *testing.T) {
	assert := assert.New(t)

	ctx := NewDB().WithConn(Default())
	ctx.err = fmt.Errorf("test error")
	inv := ctx.Invoke()
	assert.NotNil(inv.check())
}
