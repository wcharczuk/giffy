package spiffy

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

type myStruct struct {
	PrimaryKeyCol     int    `json:"pk" db:"primary_key_column,pk,serial"`
	InferredName      string `json:"normal"`
	Excluded          string `json:"-" db:"-"`
	NullableCol       string `json:"not_nullable" db:"nullable,nullable"`
	InferredWithFlags string `db:",readonly"`
}

func (m myStruct) TableName() string {
	return "my_struct"
}

func TestGetColumns(t *testing.T) {
	a := assert.New(t)

	emptyColumnCollection := ColumnCollection{}
	firstOrDefaultNil := emptyColumnCollection.FirstOrDefault()
	a.Nil(firstOrDefaultNil)

	obj := myStruct{}
	meta := NewColumnCollectionFromInstance(obj)

	a.NotNil(meta.Columns())
	a.NotEmpty(meta.Columns())

	a.Equal(4, meta.Len())

	readOnlyColumns := meta.ReadOnly()
	a.Len(readOnlyColumns.Columns(), 1)

	firstOrDefault := meta.FirstOrDefault()
	a.NotNil(firstOrDefault)

	firstCol := meta.FirstOrDefault()
	a.Equal("my_struct", firstCol.TableName)
	a.Equal("PrimaryKeyCol", firstCol.FieldName)
	a.Equal("primary_key_column", firstCol.ColumnName)
	a.True(firstCol.IsPrimaryKey)
	a.True(firstCol.IsSerial)
	a.False(firstCol.IsNullable)
	a.False(firstCol.IsReadOnly)

	secondCol := meta.Columns()[1]
	a.Equal("inferredname", secondCol.ColumnName)
	a.False(secondCol.IsPrimaryKey)
	a.False(secondCol.IsSerial)
	a.False(secondCol.IsNullable)
	a.False(secondCol.IsReadOnly)

	thirdCol := meta.Columns()[2]
	a.Equal("nullable", thirdCol.ColumnName)
	a.False(thirdCol.IsPrimaryKey)
	a.False(thirdCol.IsSerial)
	a.True(thirdCol.IsNullable)
	a.False(thirdCol.IsReadOnly)

	fourthCol := meta.Columns()[3]
	a.Equal("inferredwithflags", fourthCol.ColumnName)
	a.False(fourthCol.IsPrimaryKey)
	a.False(fourthCol.IsSerial)
	a.False(fourthCol.IsNullable)
	a.True(fourthCol.IsReadOnly)
}
