package reflectutil

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
)

// PatchStrings options.
const (
	// FieldTagEnv is the struct tag for what environment variable to use to populate a field.
	FieldTagEnv = "env"
	// FieldFlagCSV is a field tag flag (say that 10 times fast).
	FieldFlagCSV = "csv"
	// FieldFlagBase64 is a field tag flag (say that 10 times fast).
	FieldFlagBase64 = "base64"
	// FieldFlagBytes is a field tag flag (say that 10 times fast).
	FieldFlagBytes = "bytes"
)

// PatchStringer is a type that handles unmarshalling a map of strings into itself.
type PatchStringer interface {
	PatchStrings(map[string]string) error
}

// PatchStrings patches an object with a given map of data matched with tags of a given name.
func PatchStrings(tagName string, data map[string]string, obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.New(r)
		}
	}()

	// check if the type implements marshaler.
	if typed, isTyped := obj.(PatchStringer); isTyped {
		return typed.PatchStrings(data)
	}

	objMeta := reflectType(obj)
	objValue := reflectValue(obj)

	typeDuration := reflect.TypeOf(time.Duration(time.Nanosecond))

	var field reflect.StructField
	var fieldType reflect.Type
	var fieldValue reflect.Value
	var tag string
	var pieces []string
	var dataField string
	var dataValue string
	var dataFieldValue interface{}
	var hasDataValue bool

	var isCSV bool
	var isBytes bool
	var isBase64 bool
	var assigned bool

	for x := 0; x < objMeta.NumField(); x++ {
		isCSV = false
		isBytes = false
		isBase64 = false

		field = objMeta.Field(x)
		fieldValue = objValue.FieldByName(field.Name)

		// Treat structs as nested values.
		if field.Type.Kind() == reflect.Struct {
			if err = PatchStrings(tagName, data, objValue.Field(x).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		tag = field.Tag.Get(tagName)
		if len(tag) > 0 {
			pieces = strings.Split(tag, ",")
			dataField = pieces[0]
			if len(pieces) > 1 {
				for y := 1; y < len(pieces); y++ {
					if pieces[y] == FieldFlagCSV {
						isCSV = true
					} else if pieces[y] == FieldFlagBase64 {
						isBase64 = true
					} else if pieces[y] == FieldFlagBytes {
						isBytes = true
					}
				}
			}

			dataValue, hasDataValue = data[dataField]
			if !hasDataValue {
				continue
			}

			if isCSV {
				dataFieldValue = strings.Split(dataValue, ",")
			} else if isBase64 {
				dataFieldValue, err = base64.StdEncoding.DecodeString(dataValue)
				if err != nil {
					return
				}
			} else if isBytes {
				dataFieldValue = []byte(dataValue)
			} else {
				// figure out the rootmost type (i.e. deref ****ptr etc.)
				fieldType = followType(field.Type)
				switch fieldType {
				case typeDuration:
					dataFieldValue, err = time.ParseDuration(dataValue)
					if err != nil {
						err = exception.New(err)
						return
					}
				default:
					switch fieldType.Kind() {
					case reflect.Bool:
						if hasDataValue {
							dataFieldValue = mustParseBool(dataValue)
						} else {
							continue
						}
					case reflect.Float32:
						dataFieldValue, err = strconv.ParseFloat(dataValue, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Float64:
						dataFieldValue, err = strconv.ParseFloat(dataValue, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int8:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int16:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 16)
						if err != nil {
							return exception.New(err)
						}
					case reflect.Int32:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Int64:
						dataFieldValue, err = strconv.ParseInt(dataValue, 10, 64)
						if err != nil {
							return exception.New(err)
						}
					case reflect.Uint8:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint16:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 8)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint32:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 32)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint64:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.Uint, reflect.Uintptr:
						dataFieldValue, err = strconv.ParseUint(dataValue, 10, 64)
						if err != nil {
							err = exception.New(err)
							return
						}
					case reflect.String:
						dataFieldValue = dataValue
					default:
						err = exception.New("map strings into; unhandled assignment").WithMessagef("type %s", fieldType.String())
						return
					}
				}
			}

			value := reflectValue(dataFieldValue)
			if !value.IsValid() {
				err = exception.New("invalid value").WithMessagef("%s `%s`", objMeta.Name(), field.Name)
				return
			}

			assigned, err = tryAssignment(fieldValue, value)
			if err != nil {
				return
			}
			if !assigned {
				err = exception.New("cannot set field").WithMessagef("%s `%s`", objMeta.Name(), field.Name)
				return
			}
		}
	}
	return nil
}

func followType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	return t
}

func reflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func reflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func mustParseBool(str string) bool {
	strLower := strings.ToLower(str)
	switch strLower {
	case "true", "1", "yes":
		return true
	}
	return false
}

func parseBool(str string) (bool, error) {
	strLower := strings.ToLower(str)
	switch strLower {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	}
	return false, exception.New("invalid bool value").WithMessage(str)
}
