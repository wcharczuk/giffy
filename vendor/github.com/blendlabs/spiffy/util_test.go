package spiffy

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestMakeCsvTokens(t *testing.T) {
	a := assert.New(t)

	one := paramTokensCSV(1)
	two := paramTokensCSV(2)
	three := paramTokensCSV(3)

	a.Equal("$1", one)
	a.Equal("$1,$2", two)
	a.Equal("$1,$2,$3", three)
}

func TestReflectSliceType(t *testing.T) {
	assert := assert.New(t)

	objects := []benchObj{
		{}, {}, {},
	}

	ot := reflectSliceType(objects)
	assert.Equal("benchObj", ot.Name())
}

func TestMakeSliceOfType(t *testing.T) {
	a := assert.New(t)
	tx, txErr := Default().Begin()
	a.Nil(txErr)
	defer func() {
		a.Nil(tx.Rollback())
	}()

	seedErr := seedObjects(10, tx)
	a.Nil(seedErr)

	myType := reflectType(benchObj{})
	sliceOfT, castOk := makeSliceOfType(myType).(*[]benchObj)
	a.True(castOk)

	allErr := Default().GetAllInTx(sliceOfT, tx)
	a.Nil(allErr)
	a.NotEmpty(*sliceOfT)
}

type SimpleType struct {
	ID   int
	Name string
}

type SimpleTypeWithName struct {
	ID   int
	Name string
}

func (st SimpleTypeWithName) TableName() string {
	return "not_simple_type_with_name"
}

func TestTableName(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("simpletype", TableName(SimpleType{}))
	assert.Equal("simpletype", TableName(&SimpleType{}))
	assert.Equal("not_simple_type_with_name", TableName(SimpleTypeWithName{}))
	assert.Equal("not_simple_type_with_name", TableName(&SimpleTypeWithName{}))
}
