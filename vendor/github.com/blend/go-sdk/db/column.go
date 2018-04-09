package db

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/blend/go-sdk/exception"
)

// --------------------------------------------------------------------------------
// Column
// --------------------------------------------------------------------------------

// NewColumnFromFieldTag reads the contents of a field tag, ex: `json:"foo" db:"bar,isprimarykey,isserial"
func NewColumnFromFieldTag(field reflect.StructField) *Column {
	db := field.Tag.Get("db")
	if db != "-" {
		col := Column{}
		col.FieldName = field.Name
		col.ColumnName = strings.ToLower(field.Name)
		col.FieldType = field.Type
		if db != "" {
			pieces := strings.Split(db, ",")

			if !strings.HasPrefix(db, ",") {
				col.ColumnName = pieces[0]
			}

			if len(pieces) >= 1 {
				args := strings.Join(pieces[1:], ",")
				col.IsPrimaryKey = strings.Contains(strings.ToLower(args), "pk")
				col.IsSerial = strings.Contains(strings.ToLower(args), "serial")
				col.IsNullable = strings.Contains(strings.ToLower(args), "nullable")
				col.IsReadOnly = strings.Contains(strings.ToLower(args), "readonly")
				col.IsJSON = strings.Contains(strings.ToLower(args), "json")
			}
		}
		return &col
	}

	return nil
}

// Column represents a single field on a struct that is mapped to the database.
type Column struct {
	TableName    string
	FieldName    string
	FieldType    reflect.Type
	ColumnName   string
	Index        int
	IsPrimaryKey bool
	IsSerial     bool
	IsNullable   bool
	IsReadOnly   bool
	IsJSON       bool
}

// SetValue sets the field on a database mapped object to the instance of `value`.
func (c Column) SetValue(object interface{}, value interface{}) error {
	objValue := reflectValue(object)
	field := objValue.FieldByName(c.FieldName)
	fieldType := field.Type()
	if !field.CanSet() {
		return exception.New("hit a field we can't set: '" + c.FieldName + "', did you forget to pass the object as a reference?")
	}

	valueReflected := reflectValue(value)
	if !valueReflected.IsValid() {
		return nil
	}

	if c.IsJSON {
		valueAsString, ok := valueReflected.Interface().(string)
		if ok && len(valueAsString) != 0 {
			fieldAddr := field.Addr().Interface()
			jsonErr := json.Unmarshal([]byte(valueAsString), fieldAddr)
			if jsonErr != nil {
				return exception.Wrap(jsonErr)
			}
			field.Set(reflect.ValueOf(fieldAddr).Elem())
		}
		return nil
	}

	if valueReflected.Type().AssignableTo(fieldType) {
		if field.Kind() == reflect.Ptr && valueReflected.CanAddr() {
			field.Set(valueReflected.Addr())
		} else {
			field.Set(valueReflected)
		}
		return nil
	}

	if field.Kind() == reflect.Ptr {
		if valueReflected.CanAddr() {
			if fieldType.Elem() == valueReflected.Type() {
				field.Set(valueReflected.Addr())
			} else {
				convertedValue := valueReflected.Convert(fieldType.Elem())
				if convertedValue.CanAddr() {
					field.Set(convertedValue.Addr())
				}
			}
			return nil
		}

		return exception.Newf("Cannot take address of value: %#v", value)
	}

	convertedValue := valueReflected.Convert(fieldType)
	field.Set(convertedValue)

	return nil
}

// GetValue returns the value for a column on a given database mapped object.
func (c Column) GetValue(object DatabaseMapped) interface{} {
	value := reflectValue(object)
	valueField := value.Field(c.Index)
	return valueField.Interface()
}
