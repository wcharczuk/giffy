package db

import (
	"database/sql"
	"reflect"

	"github.com/blend/go-sdk/exception"
)

// PopulateByName sets the values of an object from the values of a sql.Rows object using column names.
func PopulateByName(object interface{}, row *sql.Rows, cols *ColumnCollection) error {
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
