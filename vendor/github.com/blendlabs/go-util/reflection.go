package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blendlabs/go-exception"
)

// FollowValuePointer derefs a reflectValue until it isn't a pointer, but will preseve it's nilness.
func FollowValuePointer(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	val := v
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Interface()
}

// FollowValue derefs a value until it isn't a pointer or an interface.
func FollowValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// ReflectValue returns the integral reflect.Value for an object.
func ReflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// ReflectType returns the integral type for an object.
func ReflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

// MakeNew returns a new instance of a reflect.Type.
func MakeNew(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

// MakeSliceOfType returns a new slice of a given reflect.Type.
func MakeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

// TypeName returns the string type name for an object's integral type.
func TypeName(obj interface{}) string {
	return ReflectType(obj).Name()
}

// GetValueByName returns a value for a given struct field by name.
func GetValueByName(target interface{}, fieldName string) interface{} {
	targetValue := ReflectValue(target)
	field := targetValue.FieldByName(fieldName)
	return field.Interface()
}

// GetFieldByNameOrJSONTag returns a value for a given struct field by name or by json tag name.
func GetFieldByNameOrJSONTag(targetValue reflect.Type, fieldName string) *reflect.StructField {
	for index := 0; index < targetValue.NumField(); index++ {
		field := targetValue.Field(index)

		if field.Name == fieldName {
			return &field
		}
		tag := field.Tag
		jsonTag := tag.Get("json")
		if strings.Contains(jsonTag, fieldName) {
			return &field
		}
	}

	return nil
}

// SetValueByName sets a value on an object by its field name.
func SetValueByName(target interface{}, fieldName string, fieldValue interface{}) error {
	targetValue := ReflectValue(target)
	targetType := ReflectType(target)
	relevantField := GetFieldByNameOrJSONTag(targetType, fieldName)

	if relevantField == nil {
		return exception.New(fmt.Sprintf("Invalid field for %s : `%s`", targetType.Name(), fieldName))
	}

	field := targetValue.FieldByName(relevantField.Name)
	fieldType := field.Type()
	if field.CanSet() {
		valueReflected := ReflectValue(fieldValue)
		if valueReflected.IsValid() {
			if valueReflected.Type().AssignableTo(fieldType) {
				if field.Kind() == reflect.Ptr && valueReflected.CanAddr() {
					field.Set(valueReflected.Addr())
				} else {
					field.Set(valueReflected)
				}
			} else {
				if field.Kind() == reflect.Ptr {
					if valueReflected.CanAddr() {
						convertedValue := valueReflected.Convert(fieldType.Elem())
						if convertedValue.CanAddr() {
							field.Set(convertedValue.Addr())
						}
					}
				} else {
					convertedValue := valueReflected.Convert(fieldType)
					field.Set(convertedValue)
				}
			}
		} else {
			return exception.New(fmt.Sprintf("Invalid field for %s : `%s`", targetType.Name(), fieldName))
		}
	} else {
		return exception.New(fmt.Sprintf("Cannot set field for %s : `%s`", targetType.Name(), fieldName))
	}
	return nil
}

// PatchObject updates an object based on a map of field names to values.
func PatchObject(obj interface{}, patchValues map[string]interface{}) error {
	for key, value := range patchValues {
		err := SetValueByName(obj, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecomposeToPostData dumps an object to a slice of key value tuples representing field name as form value and string value of field.
func DecomposeToPostData(object interface{}) []KeyValuePairOfString {
	kvps := []KeyValuePairOfString{}

	objType := ReflectType(object)
	objValue := ReflectValue(object)

	numberOfFields := objType.NumField()
	for index := 0; index < numberOfFields; index++ {
		field := objType.Field(index)
		valueField := objValue.Field(index)

		kvp := KeyValuePairOfString{}

		if !field.Anonymous {
			tag := field.Tag.Get("json")
			if len(tag) != 0 {
				if strings.Contains(tag, ",") {
					parts := strings.Split(tag, ",")
					kvp.Key = parts[0]
				} else {
					kvp.Key = tag
				}
			} else {
				kvp.Key = field.Name
			}

			if field.Type.Kind() == reflect.Slice {
				//do something special
				for subIndex := 0; subIndex < valueField.Len(); subIndex++ {
					itemAtIndex := valueField.Index(subIndex).Interface()
					for _, prop := range DecomposeToPostData(itemAtIndex) {
						if len(prop.Value) != 0 { //this is a gutcheck, it shouldn't be needed
							ikvp := KeyValuePairOfString{}
							ikvp.Key = fmt.Sprintf("%s[%d].%s", kvp.Key, subIndex, prop.Key)
							ikvp.Value = prop.Value
							kvps = append(kvps, ikvp)
						}
					}
				}
			} else {
				value := FollowValuePointer(valueField)
				if value != nil {
					kvp.Value = fmt.Sprintf("%v", value)
					if len(kvp.Value) != 0 {
						kvps = append(kvps, kvp)
					}
				}
			}
		}
	}

	return kvps
}

// DecomposeToPostDataAsJSON returns an array of KeyValuePairOfString for an object.
func DecomposeToPostDataAsJSON(object interface{}) []KeyValuePairOfString {
	kvps := []KeyValuePairOfString{}

	objType := ReflectType(object)
	objValue := ReflectValue(object)

	numberOfFields := objType.NumField()
	for index := 0; index < numberOfFields; index++ {
		field := objType.Field(index)
		valueField := objValue.Field(index)

		kvp := KeyValuePairOfString{}

		if !field.Anonymous {
			tag := field.Tag.Get("json")
			if len(tag) != 0 {
				if strings.Contains(tag, ",") {
					parts := strings.Split(tag, ",")
					kvp.Key = parts[0]
				} else {
					kvp.Key = tag
				}
			} else {
				kvp.Key = field.Name
			}

			valueDereferenced := FollowValue(valueField)
			value := FollowValuePointer(valueField)
			if value != nil {
				if valueDereferenced.Kind() == reflect.Slice || valueDereferenced.Kind() == reflect.Map {
					kvp.Value = SerializeJSON(value)
				} else {
					kvp.Value = fmt.Sprintf("%v", value)
				}
			}

			if len(kvp.Value) != 0 {
				kvps = append(kvps, kvp)
			}
		}
	}

	return kvps
}

// checks if a value is a zero value or its types default value
func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// Given a the name of a type variable, determines if the variable is exported
// by checking if first variable is capitalilzed
func isExported(fieldName string) bool {
	return fieldName != "" && strings.ToUpper(fieldName)[0] == fieldName[0]
}

// CoalesceFields merges non-zero fields into destination fields marked with the `coalesce:...` struct field tag.
func CoalesceFields(object interface{}) {
	objectValue := ReflectValue(object)
	objectType := ReflectType(object)
	if objectType.Kind() == reflect.Struct {
		numberOfFields := objectValue.NumField()
		for index := 0; index < numberOfFields; index++ {
			field := objectType.Field(index)
			fieldValue := objectValue.Field(index)
			// only alter the field if it is exported (uppercase variable name) and is not already a non-zero value
			if isExported(field.Name) && isZero(fieldValue) {
				alternateFieldNames := strings.Split(field.Tag.Get("coalesce"), ",")

				// find the first non-zero value in the list of backup values
				for j := 0; j < len(alternateFieldNames); j++ {
					alternateFieldName := alternateFieldNames[j]
					alternateValue := objectValue.FieldByName(alternateFieldName)
					// will panic if trying to set a non-exported value or a zero value, so ignore those
					if isExported(alternateFieldName) && !isZero(alternateValue) {
						fieldValue.Set(alternateValue)
						break
					}
				}
			}
			// recurse, in case nested values of this field need to be set as well
			if isExported(field.Name) && !isZero(fieldValue) {
				CoalesceFields(fieldValue.Addr().Interface())
			}
		}
	} else if objectType.Kind() == reflect.Array || objectType.Kind() == reflect.Slice {
		arrayLength := objectValue.Len()
		for i := 0; i < arrayLength; i++ {
			CoalesceFields(objectValue.Index(i).Addr().Interface())
		}
	}
}
