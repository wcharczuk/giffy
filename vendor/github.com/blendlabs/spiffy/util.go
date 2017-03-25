package spiffy

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/blendlabs/go-exception"
	util "github.com/blendlabs/go-util"
)

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// OptionalTx returns the first of a variadic set of txs.
// It is useful if you want to have a tx an optional parameter.
func OptionalTx(txs ...*sql.Tx) *sql.Tx {
	if len(txs) > 0 {
		return txs[0]
	}
	return nil
}

// Tx is an alias for OptionalTx
func Tx(txs ...*sql.Tx) *sql.Tx {
	return OptionalTx(txs...)
}

// TableName returns the table name for a given reflect.Type by instantiating it and calling o.TableName().
// The type must implement DatabaseMapped or an exception will be returned.
func TableName(t reflect.Type) (string, error) {
	i, err := makeNewDatabaseMapped(t)
	if err == nil {
		return i.TableName(), nil
	}
	return "", err
}

// --------------------------------------------------------------------------------
// String Utility Methods
// --------------------------------------------------------------------------------

// HasPrefixCaseInsensitive returns if a corpus has a prefix regardless of casing.
func HasPrefixCaseInsensitive(corpus, prefix string) bool {
	return util.String.HasPrefixCaseInsensitive(corpus, prefix)
}

// HasSuffixCaseInsensitive returns if a corpus has a suffix regardless of casing.
func HasSuffixCaseInsensitive(corpus, suffix string) bool {
	return util.String.HasSuffixCaseInsensitive(corpus, suffix)
}

// CaseInsensitiveEquals compares two strings regardless of case.
func CaseInsensitiveEquals(a, b string) bool {
	return util.String.CaseInsensitiveEquals(a, b)
}

// CSV returns a csv from an array.
func CSV(names []string) string {
	return strings.Join(names, ",")
}

// --------------------------------------------------------------------------------
// Internal / Reflection Utility Methods
// --------------------------------------------------------------------------------

// AsPopulatable casts an object as populatable.
func asPopulatable(object interface{}) Populatable {
	return object.(Populatable)
}

// isPopulatable returns if an object is populatable
func isPopulatable(object interface{}) bool {
	_, isPopulatable := object.(Populatable)
	return isPopulatable
}

// reflectValue returns the reflect.Value for an object following pointers.
func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// reflectType retruns the reflect.Type for an object following pointers.
func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

// reflectSliceType returns the inner type of a slice following pointers.
func reflectSliceType(collection interface{}) reflect.Type {
	v := reflect.ValueOf(collection)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Len() == 0 {
		t := v.Type()
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		return t
	}
	v = v.Index(0)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v.Type()
}

// makeWhereClause returns the sql `where` clause for a column collection, starting at a given index (used in sql $1 parameterization).
func makeWhereClause(pks *ColumnCollection, startAt int) string {
	whereClause := " WHERE "
	for i, pk := range pks.Columns() {
		whereClause = whereClause + fmt.Sprintf("%s = %s", pk.ColumnName, "$"+strconv.Itoa(i+startAt))
		if i < (pks.Len() - 1) {
			whereClause = whereClause + " AND "
		}
	}

	return whereClause
}

// paramTokensCSV returns a csv token string in the form "$1,$2,$3...$N"
func paramTokensCSV(num int) string {
	str := ""
	for i := 1; i <= num; i++ {
		str = str + fmt.Sprintf("$%d", i)
		if i != num {
			str = str + ","
		}
	}
	return str
}

// makeNewDatabaseMapped returns a new instance of a database mapped type.
func makeNewDatabaseMapped(t reflect.Type) (DatabaseMapped, error) {
	newInterface := reflect.New(t).Interface()
	if typed, isTyped := newInterface.(DatabaseMapped); isTyped {
		return typed.(DatabaseMapped), nil
	}
	return nil, exception.Newf("`%s` does not implement DatabaseMapped.", t.Name())
}

// makeNew creates a new object.
func makeNew(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

func makeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}
