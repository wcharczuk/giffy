package spiffy

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestMakeCsvTokens(t *testing.T) {
	a := assert.New(t)

	one := ParamTokensCSV(1)
	two := ParamTokensCSV(2)
	three := ParamTokensCSV(3)

	a.Equal("$1", one)
	a.Equal("$1,$2", two)
	a.Equal("$1,$2,$3", three)
}

func TestMakeSliceOfType(t *testing.T) {
	a := assert.New(t)
	tx, txErr := DefaultDb().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	myType := reflectType(benchObj{})
	sliceOfT, castOk := makeSliceOfType(myType).(*[]benchObj)
	a.True(castOk)

	allErr := DefaultDb().GetAllInTransaction(sliceOfT, tx)
	a.Nil(allErr)
	a.NotEmpty(*sliceOfT)
}
