package web

import (
	"reflect"
	"strconv"
	"strings"

	exception "github.com/blendlabs/go-exception"
)

// ReflectValue returns the integral reflect.Value for an object.
func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// ReflectType returns the integral type for an object.
func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

// getFieldByNameOrJSONTag returns a value for a given struct field by name or by json tag name.
func getFieldByNameOrTag(targetValue reflect.Type, tagName, fieldName string) *reflect.StructField {
	for index := 0; index < targetValue.NumField(); index++ {
		field := targetValue.Field(index)

		if field.Name == fieldName {
			return &field
		}
		tag := field.Tag
		jsonTag := tag.Get(tagName)
		if strings.Contains(jsonTag, fieldName) {
			return &field
		}
	}

	return nil
}

// GetValueByName returns a value for a given struct field by name.
func getValueByName(target interface{}, fieldName string) interface{} {
	targetValue := reflectValue(target)
	field := targetValue.FieldByName(fieldName)
	return field.Interface()
}

// SetValueByName sets a value on an object by its field name.
func setValueByName(target interface{}, fieldName string, fieldValue interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.Newf("Error setting field: %v", r)
		}
	}()
	typeCheck := reflect.TypeOf(target)
	if typeCheck.Kind() != reflect.Ptr {
		return exception.New("Cannot modify non-pointer target")
	}

	targetValue := reflectValue(target)
	targetType := reflectType(target)
	relevantField, hasField := targetType.FieldByName(fieldName)

	if !hasField {
		return exception.Newf("Field not found  %s.%s", targetType.Name(), fieldName)
	}

	field := targetValue.FieldByName(relevantField.Name)
	fieldType := field.Type()
	if !field.CanSet() {
		return exception.Newf("Cannot set field %s", fieldName)
	}

	valueReflected := reflectValue(fieldValue)
	if !valueReflected.IsValid() {
		return exception.New("Reflected value is invalid, cannot continue.")
	}

	if valueReflected.Type().AssignableTo(fieldType) {
		field.Set(valueReflected)
		return nil
	}

	if field.Kind() == reflect.Ptr {
		if valueReflected.CanAddr() {
			convertedValue := valueReflected.Convert(fieldType.Elem())
			if convertedValue.CanAddr() {
				field.Set(convertedValue.Addr())
				return nil
			}
		}
		return exception.New("Cannot take address of value for assignment to field pointer")
	}

	if fieldAsString, isString := valueReflected.Interface().(string); isString {
		var parsedValue reflect.Value
		handledType := true
		switch fieldType.Kind() {
		case reflect.Int:
			intValue, err := strconv.Atoi(fieldAsString)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(intValue)
		case reflect.Int64:
			int64Value, err := strconv.ParseInt(fieldAsString, 10, 64)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(int64Value)
		case reflect.Uint16:
			intValue, err := strconv.Atoi(fieldAsString)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(uint16(intValue))
		case reflect.Uint: //a.k.a. uint32
			intValue, err := strconv.Atoi(fieldAsString)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(uint(intValue))
		case reflect.Uint32:
			intValue, err := strconv.Atoi(fieldAsString)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(uint32(intValue))
		case reflect.Uint64:
			intValue, err := strconv.Atoi(fieldAsString)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(uint64(intValue))
		case reflect.Float32:
			floatValue, err := strconv.ParseFloat(fieldAsString, 32)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(floatValue)
		case reflect.Float64:
			floatValue, err := strconv.ParseFloat(fieldAsString, 64)
			if err != nil {
				return exception.Wrap(err)
			}
			parsedValue = reflect.ValueOf(floatValue)
		default:
			handledType = false
		}
		if handledType {
			field.Set(parsedValue)
			return nil
		}
	}

	convertedValue := valueReflected.Convert(fieldType)
	if convertedValue.IsValid() && convertedValue.Type().AssignableTo(fieldType) {
		field.Set(convertedValue)
		return nil
	}

	return exception.New("Couldnt set field %s.%s", targetType.Name(), fieldName)
}
