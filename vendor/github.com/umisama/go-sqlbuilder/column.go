package sqlbuilder

import (
	sqldriver "database/sql/driver"
	"reflect"
	"time"
)

// ColumnConfig represents a config for table's column.
// This has a name, data type and some options.
type ColumnConfig interface {
	serializable

	toColumn(Table) Column
	Name() string
	Type() ColumnType
	Option() *ColumnOption
}

// ColumnType reprecents a type of column.
// Dialects handle this for know column options.
type ColumnType int

const (
	ColumnTypeAny ColumnType = iota
	ColumnTypeInt
	ColumnTypeString
	ColumnTypeDate
	ColumnTypeFloat
	ColumnTypeBool
	ColumnTypeBytes
)

func (t ColumnType) String() string {
	switch t {
	case ColumnTypeInt:
		return "int"
	case ColumnTypeString:
		return "string"
	case ColumnTypeDate:
		return "date"
	case ColumnTypeFloat:
		return "float"
	case ColumnTypeBool:
		return "bool"
	case ColumnTypeBytes:
		return "bytes"
	case ColumnTypeAny:
		return "any"
	}
	panic(newError("unknown columnType"))
}

func (t ColumnType) CapableTypes() []reflect.Type {
	switch t {
	case ColumnTypeInt:
		return []reflect.Type{
			reflect.TypeOf(int(0)),
			reflect.TypeOf(int8(0)),
			reflect.TypeOf(int16(0)),
			reflect.TypeOf(int32(0)),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf(uint(0)),
			reflect.TypeOf(uint8(0)),
			reflect.TypeOf(uint16(0)),
			reflect.TypeOf(uint32(0)),
			reflect.TypeOf(uint64(0)),
		}
	case ColumnTypeString:
		return []reflect.Type{
			reflect.TypeOf(""),
		}
	case ColumnTypeDate:
		return []reflect.Type{
			reflect.TypeOf(time.Time{}),
		}
	case ColumnTypeFloat:
		return []reflect.Type{
			reflect.TypeOf(float32(0)),
			reflect.TypeOf(float64(0)),
		}
	case ColumnTypeBool:
		return []reflect.Type{
			reflect.TypeOf(bool(true)),
		}
	case ColumnTypeBytes:
		return []reflect.Type{
			reflect.TypeOf([]byte{}),
		}
	case ColumnTypeAny:
		return []reflect.Type{} // but accept all types
	}
	return []reflect.Type{}
}

// ColumnOption represents option for a column. ex: primary key.
type ColumnOption struct {
	PrimaryKey    bool
	NotNull       bool
	Unique        bool
	AutoIncrement bool
	Size          int
	SqlType       string
	Default       interface{}
}

// ColumnList represents list of Column.
type ColumnList []Column

// Column represents a table column.
type Column interface {
	serializable

	column_name() string
	config() ColumnConfig
	acceptType(interface{}) bool

	// As creates Column alias.
	As(alias string) Column

	// Eq creates Condition for "column==right".  Type for right is column's one or other Column.
	Eq(right interface{}) Condition

	// NotEq creates Condition for "column<>right".  Type for right is column's one or other Column.
	NotEq(right interface{}) Condition

	// GtEq creates Condition for "column>right".  Type for right is column's one or other Column.
	Gt(right interface{}) Condition

	// GtEq creates Condition for "column>=right".  Type for right is column's one or other Column.
	GtEq(right interface{}) Condition

	// Lt creates Condition for "column<right".  Type for right is column's one or other Column.
	Lt(right interface{}) Condition

	// LtEq creates Condition for "column<=right".  Type for right is column's one or other Column.
	LtEq(right interface{}) Condition

	// Like creates Condition for "column LIKE right".  Type for right is column's one or other Column.
	Like(right string) Condition

	// Between creates Condition for "column BETWEEN lower AND higher".  Type for lower/higher is int or time.Time.
	Between(lower, higher interface{}) Condition

	// In creates Condition for "column IN (values[0], values[1] ...)".  Type for values is column's one or other Column.
	In(values ...interface{}) Condition

	// IntersectJSON creates Condition for "column @> value".  The value is a JSON string.
	IntersectJSON(data string) Condition

	// NotIn creates Condition for "column NOT IN (values[0], values[1] ...)".  Type for values is column's one or other Column.
	NotIn(values ...interface{}) Condition
}

type aliasedColumn interface {
	Column
	column_alias() string
	source() Column
}

type columnConfigImpl struct {
	name string
	typ  ColumnType
	opt  *ColumnOption
}

func (c *columnConfigImpl) Name() string {
	return c.name
}

func (c *columnConfigImpl) Type() ColumnType {
	return c.typ
}

func (c *columnConfigImpl) Option() *ColumnOption {
	if c.opt == nil {
		return &ColumnOption{}
	}
	return c.opt
}

func (m *columnConfigImpl) toColumn(table Table) Column {
	return &columnImpl{
		m, table,
	}
}

func (m *columnConfigImpl) serialize(bldr *builder) {
	bldr.Append(dialect().QuoteField(m.name))
	return
}

type columnImpl struct {
	*columnConfigImpl
	table Table
}

func (m *columnImpl) column_name() string {
	return m.name
}

func (m *columnImpl) config() ColumnConfig {
	return m.columnConfigImpl
}

func (m *columnImpl) acceptType(val interface{}) bool {
	lit, ok := val.(literal)
	if !ok || lit == nil {
		return false
	}
	if lit.Raw() == nil {
		return !m.opt.NotNull
	}
	if m.Type() == ColumnTypeAny {
		return true
	}
	if _, ok := lit.Raw().(sqldriver.Valuer); ok {
		return true
	}

	valt := reflect.TypeOf(lit.Raw())
	for _, t := range m.typ.CapableTypes() {
		if t == valt {
			return true
		}
	}
	return false
}

func (m *columnImpl) serialize(bldr *builder) {
	if m == Star {
		bldr.Append("*")
	} else {
		bldr.Append(dialect().QuoteField(m.table.Name()) + "." + dialect().QuoteField(m.name))
	}
	return
}

func (m *columnImpl) As(alias string) Column {
	return &aliasColumn{
		column: m,
		alias:  alias,
	}
}

// AnyColumn creates config for any types.
func AnyColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeAny, opt)
}

// IntColumn creates config for INTEGER type column.
func IntColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeInt, opt)
}

// StringColumn creates config for TEXT or VARCHAR type column.
func StringColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeString, opt)
}

// DateColumn creates config for DATETIME type column.
func DateColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeDate, opt)
}

// FloatColumn creates config for REAL or FLOAT type column.
func FloatColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeFloat, opt)
}

// BoolColumn creates config for BOOLEAN type column.
func BoolColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeBool, opt)
}

// BytesColumn creates config for BLOB type column.
func BytesColumn(name string, opt *ColumnOption) ColumnConfig {
	return newColumnConfigImpl(name, ColumnTypeBytes, opt)
}

func newColumnConfigImpl(name string, typ ColumnType, opt *ColumnOption) *columnConfigImpl {
	if opt == nil {
		opt = &ColumnOption{}
	}
	return &columnConfigImpl{
		name: name,
		typ:  typ,
		opt:  opt,
	}
}

func (left *columnImpl) Eq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "=")
}

func (left *columnImpl) NotEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<>")
}

func (left *columnImpl) Gt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">")
}

func (left *columnImpl) GtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">=")
}

func (left *columnImpl) Lt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<")
}

func (left *columnImpl) LtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<=")
}

func (left *columnImpl) Like(right string) Condition {
	return newBinaryOperationCondition(left, right, " LIKE ")
}

func (left *columnImpl) Between(lower, higher interface{}) Condition {
	return newBetweenCondition(left, lower, higher)
}

func (left *columnImpl) In(val ...interface{}) Condition {
	return newInCondition(left, val...)
}

func (left *columnImpl) IntersectJSON(data string) Condition {
	return newIntersectJSONCondition(left, data)
}

func (left *columnImpl) NotIn(val ...interface{}) Condition {
	return newNotInCondition(left, val...)
}

func (b ColumnList) serialize(bldr *builder) {
	first := true
	for _, column := range b {
		if column == nil {
			bldr.SetError(newError("column is not found."))
			return
		}
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		bldr.Append(dialect().QuoteField(column.column_name()))
	}
	return
}

type errorColumn struct {
	err error
}

func newErrorColumn(err error) Column {
	return &errorColumn{
		err: err,
	}
}

func (m *errorColumn) column_name() string {
	return ""
}

func (m *errorColumn) config() ColumnConfig {
	return nil
}

func (m *errorColumn) acceptType(interface{}) bool {
	return false
}

func (m *errorColumn) serialize(bldr *builder) {
	bldr.SetError(m.err)
	return
}

func (m *errorColumn) As(string) Column {
	return m
}

func (left *errorColumn) Eq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "=")
}

func (left *errorColumn) NotEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<>")
}

func (left *errorColumn) Gt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">")
}

func (left *errorColumn) GtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">=")
}

func (left *errorColumn) Lt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<")
}

func (left *errorColumn) LtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<=")
}

func (left *errorColumn) Like(right string) Condition {
	return newBinaryOperationCondition(left, right, " LIKE ")
}

func (left *errorColumn) Between(lower, higher interface{}) Condition {
	return newBetweenCondition(left, lower, higher)
}

func (left *errorColumn) In(val ...interface{}) Condition {
	return newInCondition(left, val...)
}

func (left *errorColumn) IntersectJSON(data string) Condition {
	return newIntersectJSONCondition(left, data)
}

func (left *errorColumn) NotIn(val ...interface{}) Condition {
	return newNotInCondition(left, val...)
}

type aliasColumn struct {
	column Column
	alias  string
}

func (m *aliasColumn) column_name() string {
	return m.alias
}

func (m *aliasColumn) config() ColumnConfig {
	return m.column.config()
}

func (m *aliasColumn) acceptType(val interface{}) bool {
	return m.column.acceptType(val)
}

func (m *aliasColumn) As(alias string) Column {
	return &aliasColumn{
		column: m,
		alias:  alias,
	}
}

func (m *aliasColumn) serialize(bldr *builder) {
	bldr.Append(dialect().QuoteField(m.alias))
	return
}

func (m *aliasColumn) column_alias() string {
	return m.alias
}

func (m *aliasColumn) source() Column {
	return m.column
}

func (left *aliasColumn) Eq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "=")
}

func (left *aliasColumn) NotEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<>")
}

func (left *aliasColumn) Gt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">")
}

func (left *aliasColumn) GtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">=")
}

func (left *aliasColumn) Lt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<")
}

func (left *aliasColumn) LtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<=")
}

func (left *aliasColumn) Like(right string) Condition {
	return newBinaryOperationCondition(left, right, " LIKE ")
}

func (left *aliasColumn) Between(lower, higher interface{}) Condition {
	return newBetweenCondition(left, lower, higher)
}

func (left *aliasColumn) In(val ...interface{}) Condition {
	return newInCondition(left, val...)
}

func (left *aliasColumn) IntersectJSON(data string) Condition {
	return newIntersectJSONCondition(left, data)
}

func (left *aliasColumn) NotIn(val ...interface{}) Condition {
	return newNotInCondition(left, val...)
}
