package spiffy

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestSetValue(t *testing.T) {
	a := assert.New(t)
	obj := myStruct{InferredName: "Hello."}

	var value interface{}
	value = 10
	meta := NewColumnCollectionFromInstance(obj)
	pk := meta.Columns()[0]
	a.Nil(pk.SetValue(&obj, value))
	a.Equal(10, obj.PrimaryKeyCol)
}

func TestGetValue(t *testing.T) {
	a := assert.New(t)
	obj := myStruct{PrimaryKeyCol: 5, InferredName: "Hello."}

	meta := NewColumnCollectionFromInstance(obj)
	pk := meta.PrimaryKeys().FirstOrDefault()
	value := pk.GetValue(&obj)
	a.NotNil(value)
	a.Equal(5, value)
}
