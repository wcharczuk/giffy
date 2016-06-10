package spiffy

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/blendlabs/go-exception"
)

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// AsPopulatable casts an object as populatable.
func AsPopulatable(object DatabaseMapped) Populatable {
	return object.(Populatable)
}

// IsPopulatable returns if an object is populatable
func IsPopulatable(object DatabaseMapped) bool {
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
	t := reflect.TypeOf(collection)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return t
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

// ParamTokensCSV returns a csv token string in the form "$1,$2,$3...$N"
func ParamTokensCSV(num int) string {
	str := ""
	for i := 1; i <= num; i++ {
		str = str + fmt.Sprintf("$%d", i)
		if i != num {
			str = str + ","
		}
	}
	return str
}

// TableName returns the table name for a given reflect.Type by instantiating it and calling o.TableName().
// The type must implement DatabaseMapped or an exception will be returned.
func TableName(t reflect.Type) (string, error) {
	i, err := MakeNew(t)
	if err == nil {
		return i.TableName(), nil
	}
	return "", err
}

// MakeNew returns a new instance of a database mapped type.
func MakeNew(t reflect.Type) (DatabaseMapped, error) {
	newInterface := reflect.New(t).Interface()
	if typed, isTyped := newInterface.(DatabaseMapped); isTyped {
		return typed.(DatabaseMapped), nil
	}
	return nil, exception.New("`t` does not implement DatabaseMapped.")
}

func makeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

// PopulateByName sets the values of an object from the values of a sql.Rows object using column names.
func PopulateByName(object DatabaseMapped, row *sql.Rows, cols *ColumnCollection) error {
	rowColumns, rowColumnsErr := row.Columns()

	if rowColumnsErr != nil {
		return exception.Wrap(rowColumnsErr)
	}

	var values = make([]interface{}, len(rowColumns))
	var columnLookup = cols.Lookup()

	for i, name := range rowColumns {
		if col, ok := columnLookup[name]; ok {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				values[i] = reflect.New(reflect.PtrTo(col.FieldType)).Interface()
			}
		} else {
			var value interface{}
			values[i] = &value
		}
	}

	scanErr := row.Scan(values...)

	if scanErr != nil {
		return exception.Wrap(scanErr)
	}

	for i, v := range values {
		colName := rowColumns[i]

		if field, ok := columnLookup[colName]; ok {
			err := field.SetValue(object, v)
			if err != nil {
				return exception.Wrap(err)
			}
		}
	}

	return nil
}

// PopulateInOrder sets the values of an object in order from a sql.Rows object.
// Only use this method if you're certain of the column order. It is faster than populateByName.
// Optionally if your object implements Populatable this process will be skipped completely, which is even faster.
func PopulateInOrder(object DatabaseMapped, row *sql.Rows, cols *ColumnCollection) error {
	var values = make([]interface{}, cols.Len())

	for i, col := range cols.Columns() {
		if col.FieldType.Kind() == reflect.Ptr {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				blankPtr := reflect.New(reflect.PtrTo(col.FieldType))
				if blankPtr.CanAddr() {
					values[i] = blankPtr.Addr()
				} else {
					values[i] = blankPtr.Interface()
				}
			}
		} else {
			if col.IsJSON {
				str := ""
				values[i] = &str
			} else {
				values[i] = reflect.New(reflect.PtrTo(col.FieldType)).Interface()
			}
		}
	}

	scanErr := row.Scan(values...)

	if scanErr != nil {
		return exception.Wrap(scanErr)
	}

	columns := cols.Columns()
	for i, v := range values {
		field := columns[i]
		err := field.SetValue(object, v)
		if err != nil {
			return exception.Wrap(err)
		}
	}

	return nil
}

// CSV returns a csv from an array.
func CSV(names []string) string {
	return strings.Join(names, ",")
}
