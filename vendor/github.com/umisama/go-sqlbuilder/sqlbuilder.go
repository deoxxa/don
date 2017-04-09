// Package sqlbuilder is a SQL-query builder for golang.  This supports you using relational database with more readable and flexible code than raw SQL query string.
//
// See https://github.com/umisama/go-sqlbuilder for more infomation.
package sqlbuilder

import (
	"bytes"
	"fmt"
)

var _dialect Dialect = nil

// Star reprecents * column.
var Star Column = &columnImpl{nil, nil}

// Statement reprecents a statement(SELECT/INSERT/UPDATE and other)
type Statement interface {
	ToSql() (query string, attrs []interface{}, err error)
}

type serializable interface {
	serialize(b *builder)
}

// Dialect encapsulates behaviors that differ across SQL database.
type Dialect interface {
	QuerySuffix() string
	BindVar(i int) string
	QuoteField(field interface{}) string
	ColumnTypeToString(ColumnConfig) (string, error)
	ColumnOptionToString(*ColumnOption) (string, error)
	TableOptionToString(*TableOption) (string, error)
}

// SetDialect sets dialect for SQL server.
// Must set dialect at first.
func SetDialect(opt Dialect) {
	_dialect = opt
}

func dialect() Dialect {
	if _dialect == nil {
		panic(newError("dialect is not setted.  Call SetDialect() first."))
	}
	return _dialect
}

type builder struct {
	query *bytes.Buffer
	args  []interface{}
	err   error
}

func newBuilder() *builder {
	return &builder{
		query: bytes.NewBuffer(make([]byte, 0, 256)),
		args:  make([]interface{}, 0, 8),
		err:   nil,
	}
}

func (b *builder) Err() error {
	if b.err != nil {
		return b.err
	}
	return nil
}

func (b *builder) Query() string {
	if b.err != nil {
		return ""
	}
	return b.query.String() + dialect().QuerySuffix()
}

func (b *builder) Args() []interface{} {
	if b.err != nil {
		return []interface{}{}
	}
	return b.args
}

func (b *builder) SetError(err error) {
	if b.err != nil {
		return
	}
	b.err = err
	return
}

func (b *builder) Append(query string) {
	if b.err != nil {
		return
	}

	b.query.WriteString(query)
}

func (b *builder) AppendValue(val interface{}) {
	if b.err != nil {
		return
	}

	b.query.WriteString(dialect().BindVar(len(b.args) + 1))
	b.args = append(b.args, val)
	return
}

func (b *builder) AppendItems(parts []serializable, sep string) {
	if parts == nil {
		return
	}

	first := true
	for _, part := range parts {
		if first {
			first = false
		} else {
			b.Append(sep)
		}
		part.serialize(b)
	}
	return
}

func (b *builder) AppendItem(part serializable) {
	if part == nil {
		return
	}
	part.serialize(b)
}

type errors struct {
	fmt  string
	args []interface{}
}

func newError(fmt string, args ...interface{}) *errors {
	return &errors{
		fmt:  fmt,
		args: args,
	}
}

func (err *errors) Error() string {
	return fmt.Sprintf("sqlbuilder: "+err.fmt, err.args...)
}
