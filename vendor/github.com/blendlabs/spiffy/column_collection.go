package spiffy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

var (
	metaCacheLock sync.Mutex
	metaCache     map[string]*ColumnCollection
)

// --------------------------------------------------------------------------------
// Utility
// --------------------------------------------------------------------------------

// ColumnNames returns a csv of column names.
func ColumnNames(object DatabaseMapped) string {
	columns := CachedColumnCollectionFromInstance(object)
	return CSV(columns.ColumnNames())
}

// --------------------------------------------------------------------------------
// Column Collection
// --------------------------------------------------------------------------------
// func NewColumnCollection() *ColumnCollection { return &ColumnCollection{lookup: map[string]*Column} }

// NewColumnCollectionWithPrefix makes a new column collection with a column prefix.
func NewColumnCollectionWithPrefix(columnPrefix string) *ColumnCollection {
	return &ColumnCollection{lookup: map[string]*Column{}, columnPrefix: columnPrefix}
}

// NewColumnCollectionFromColumns creates a column lookup for a slice of columns.
func NewColumnCollectionFromColumns(columns []Column) *ColumnCollection {
	cc := ColumnCollection{columns: columns}
	lookup := make(map[string]*Column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.lookup = lookup
	return &cc
}

// Meta is an alias to CachedColumnCollectionFromInstance
func Meta(obj DatabaseMapped) *ColumnCollection {
	return CachedColumnCollectionFromInstance(obj)
}

// MakeColumnCacheKey creates a cache key for a type.
func MakeColumnCacheKey(objectType reflect.Type) string {
	if dbm, err := MakeNewDatabaseMapped(objectType); err == nil {
		return fmt.Sprintf("%s_%s", objectType.String(), dbm.TableName())
	}
	return objectType.String()
}

// CachedColumnCollectionFromInstance reflects an object instance into a new column collection.
func CachedColumnCollectionFromInstance(object interface{}) *ColumnCollection {
	objectType := reflect.TypeOf(object)
	return CachedColumnCollectionFromType(MakeColumnCacheKey(objectType), objectType)
}

func initMetaCache() {
	metaCacheLock.Lock()
	defer metaCacheLock.Unlock()

	if metaCache == nil {
		metaCache = map[string]*ColumnCollection{}
	}
}

func addMetaCacheEntry(identifier string, t reflect.Type) *ColumnCollection {
	metaCacheLock.Lock()
	defer metaCacheLock.Unlock()
	if _, ok := metaCache[identifier]; !ok {
		metadata := GenerateColumnCollectionForType(t)
		metaCache[identifier] = metadata
		return metadata
	}
	return metaCache[identifier]
}

// CachedColumnCollectionFromType reflects a reflect.Type into a column collection.
// The results of this are cached for speed.
func CachedColumnCollectionFromType(identifier string, t reflect.Type) *ColumnCollection {
	if metaCache == nil {
		initMetaCache()
	}

	if _, ok := metaCache[identifier]; !ok {
		return addMetaCacheEntry(identifier, t)
	}
	return metaCache[identifier]
}

// GenerateColumnCollectionForType reflects a new column collection from a reflect.Type.
func GenerateColumnCollectionForType(t reflect.Type) *ColumnCollection {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName, _ := TableName(t)
	numFields := t.NumField()

	var cols []Column
	for index := 0; index < numFields; index++ {
		field := t.Field(index)
		if !field.Anonymous {
			col := NewColumnFromFieldTag(field)
			if col != nil {
				col.Index = index
				col.TableName = tableName
				cols = append(cols, *col)
			}
		}
	}

	return NewColumnCollectionFromColumns(cols)
}

// ColumnCollection represents the column metadata for a given struct.
type ColumnCollection struct {
	columns      []Column
	lookup       map[string]*Column
	columnPrefix string
}

// Len returns the number of columns.
func (cc *ColumnCollection) Len() int {
	return len(cc.columns)
}

// Add adds a column.
func (cc *ColumnCollection) Add(c Column) {
	cc.columns = append(cc.columns, c)
	cc.lookup[c.ColumnName] = &c
}

// Remove removes a column (by column name) from the collection.
func (cc *ColumnCollection) Remove(columnName string) {
	var newColumns []Column
	for _, c := range cc.columns {
		if c.ColumnName != columnName {
			newColumns = append(newColumns, c)
		}
	}
	cc.columns = newColumns
	delete(cc.lookup, columnName)
}

// HasColumn returns if a column name is present in the collection.
func (cc *ColumnCollection) HasColumn(columnName string) bool {
	_, hasColumn := cc.lookup[columnName]
	return hasColumn
}

// Copy creates a new column collection instance and carries over an existing column prefix.
func (cc ColumnCollection) Copy() *ColumnCollection {
	newCC := NewColumnCollectionFromColumns(cc.columns)
	newCC.columnPrefix = cc.columnPrefix
	return newCC
}

// CopyWithColumnPrefix applies a column prefix to column names and returns a new column collection.
func (cc ColumnCollection) CopyWithColumnPrefix(prefix string) *ColumnCollection {
	newCC := NewColumnCollectionFromColumns(cc.columns)
	newCC.columnPrefix = prefix
	return newCC
}

// PrimaryKeys are columns we use as where predicates and can't update.
func (cc ColumnCollection) PrimaryKeys() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)
	newCC.columnPrefix = cc.columnPrefix

	for _, c := range cc.columns {
		if c.IsPrimaryKey {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotPrimaryKeys are columns we can update.
func (cc ColumnCollection) NotPrimaryKeys() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsPrimaryKey {
			newCC.Add(c)
		}
	}

	return newCC
}

// Serials are columns we have to return the id of.
func (cc ColumnCollection) Serials() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if c.IsSerial {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotSerials are columns we don't have to return the id of.
func (cc ColumnCollection) NotSerials() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsSerial {
			newCC.Add(c)
		}
	}

	return newCC
}

// ReadOnly are columns that we don't have to insert upon Create().
func (cc ColumnCollection) ReadOnly() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if c.IsReadOnly {
			newCC.Add(c)
		}
	}

	return newCC
}

// NotReadOnly are columns that we have to insert upon Create().
func (cc ColumnCollection) NotReadOnly() *ColumnCollection {
	newCC := NewColumnCollectionWithPrefix(cc.columnPrefix)

	for _, c := range cc.columns {
		if !c.IsReadOnly {
			newCC.Add(c)
		}
	}

	return newCC
}

// ColumnNames returns the string names for all the columns in the collection.
func (cc ColumnCollection) ColumnNames() []string {
	var names []string
	for _, c := range cc.columns {
		if len(cc.columnPrefix) != 0 {
			names = append(names, fmt.Sprintf("%s%s", cc.columnPrefix, c.ColumnName))
		} else {
			names = append(names, c.ColumnName)
		}
	}
	return names
}

// Columns returns the colummns
func (cc ColumnCollection) Columns() []Column {
	return cc.columns
}

// Lookup gets the column name lookup.
func (cc ColumnCollection) Lookup() map[string]*Column {
	if len(cc.columnPrefix) != 0 {
		lookup := map[string]*Column{}
		for key, value := range cc.lookup {
			lookup[fmt.Sprintf("%s%s", cc.columnPrefix, key)] = value
		}
		return lookup
	}
	return cc.lookup
}

// ColumnNamesFromAlias returns the string names for all the columns in the collection.
func (cc ColumnCollection) ColumnNamesFromAlias(tableAlias string) []string {
	var names []string
	for _, c := range cc.columns {
		if len(cc.columnPrefix) != 0 {
			names = append(names, fmt.Sprintf("%s.%s as %s%s", tableAlias, c.ColumnName, cc.columnPrefix, c.ColumnName))
		} else {
			names = append(names, fmt.Sprintf("%s.%s", tableAlias, c.ColumnName))
		}
	}
	return names
}

// ColumnValues returns the reflected value for all the columns on a given instance.
func (cc ColumnCollection) ColumnValues(instance interface{}) []interface{} {
	value := reflectValue(instance)

	var values []interface{}
	for _, c := range cc.columns {
		valueField := value.FieldByName(c.FieldName)
		if c.IsJSON {
			toSerialize := valueField.Interface()
			jsonBytes, _ := json.Marshal(toSerialize)
			values = append(values, string(jsonBytes))
		} else {
			values = append(values, valueField.Interface())
		}
	}
	return values
}

// FirstOrDefault returns the first column in the collection or `nil` if the collection is empty.
func (cc ColumnCollection) FirstOrDefault() *Column {
	if len(cc.columns) > 0 {
		col := cc.columns[0]
		return &col
	}
	return nil
}

// ConcatWith merges a collection with another collection.
func (cc ColumnCollection) ConcatWith(other *ColumnCollection) *ColumnCollection {
	var total []Column
	total = append(total, cc.columns...)
	total = append(total, other.columns...)
	return NewColumnCollectionFromColumns(total)
}
