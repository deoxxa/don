package dialects

import (
	"errors"
	"fmt"
	sb "github.com/umisama/go-sqlbuilder"
	"strconv"
	"time"
)

type Postgresql struct{}

func (m Postgresql) QuerySuffix() string {
	return ";"
}

func (m Postgresql) BindVar(i int) string {
	return "$" + strconv.Itoa(i)
}

func (m Postgresql) quoteField(field interface{}) (string, bool) {
	str := ""
	bracket := true
	switch t := field.(type) {
	case string:
		str = t
	case []byte:
		str = string(t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		str = fmt.Sprint(field)
	case float32, float64:
		str = fmt.Sprint(field)
	case time.Time:
		str = t.Format("2006-01-02 15:04:05")
	case bool:
		if t {
			str = "TRUE"
		} else {
			str = "FALSE"
		}
		bracket = false
	case nil:
		str = "NULL"
		bracket = false
	}
	return str, bracket
}

func (m Postgresql) QuoteField(field interface{}) string {
	str, bracket := m.quoteField(field)
	if bracket {
		str = "\"" + str + "\""
	}
	return str
}

func (m Postgresql) ColumnTypeToString(cc sb.ColumnConfig) (string, error) {
	if cc.Option().SqlType != "" {
		return cc.Option().SqlType, nil
	}

	typ := ""
	switch cc.Type() {
	case sb.ColumnTypeInt:
		if cc.Option().AutoIncrement {
			typ = "SERIAL"
		} else {
			typ = "BIGINT"
		}
	case sb.ColumnTypeString:
		typ = fmt.Sprintf("VARCHAR(%d)", cc.Option().Size)
	case sb.ColumnTypeDate:
		typ = "TIMESTAMP"
	case sb.ColumnTypeFloat:
		typ = "REAL"
	case sb.ColumnTypeBool:
		typ = "BOOLEAN"
	case sb.ColumnTypeBytes:
		typ = "BYTEA"
	}

	if typ == "" {
		return "", errors.New("dialects: unknown column type")
	} else {
		return typ, nil
	}
}

func (m Postgresql) ColumnOptionToString(co *sb.ColumnOption) (string, error) {
	opt := ""
	if co.PrimaryKey {
		opt = str_append(opt, "PRIMARY KEY")
	}
	if co.AutoIncrement {
		// do nothing
	}
	if co.NotNull {
		opt = str_append(opt, "NOT NULL")
	}
	if co.Unique {
		opt = str_append(opt, "UNIQUE")
	}
	if co.Default == nil {
		if !co.PrimaryKey {
			opt = str_append(opt, "DEFAULT NULL")
		}
	} else {
		str, bracket := m.quoteField(co.Default)
		if bracket {
			str = "'" + str + "'"
		}
		opt = str_append(opt, "DEFAULT "+str)
	}

	return opt, nil
}

func (m Postgresql) TableOptionToString(to *sb.TableOption) (string, error) {
	opt := ""
	if to.Unique != nil {
		opt = str_append(opt, m.tableOptionUnique(to.Unique))
	}

	return "", nil
}

func (m Postgresql) tableOptionUnique(op [][]string) string {
	opt := ""
	first_op := true
	for _, unique := range op {
		if first_op {
			first_op = false
		} else {
			opt += " "
		}

		opt += "UNIQUE("
		first := true
		for _, col := range unique {
			if first {
				first = false
			} else {
				opt += ", "
			}
			opt += m.QuoteField(col)
		}
		opt += ")"
	}
	return opt
}
