package dialects

import (
	"errors"
	"fmt"
	sb "github.com/umisama/go-sqlbuilder"
	"time"
)

type Sqlite struct{}

func (m Sqlite) QuerySuffix() string {
	return ";"
}

func (m Sqlite) BindVar(i int) string {
	return "?"
}

func (m Sqlite) QuoteField(field interface{}) string {
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
		return "NULL"
		bracket = false
	}
	if bracket {
		str = "\"" + str + "\""
	}
	return str
}

func (m Sqlite) ColumnTypeToString(cc sb.ColumnConfig) (string, error) {
	if cc.Option().SqlType != "" {
		return cc.Option().SqlType, nil
	}

	typ := ""
	switch cc.Type() {
	case sb.ColumnTypeInt:
		typ = "INTEGER"
	case sb.ColumnTypeString:
		typ = "TEXT"
	case sb.ColumnTypeDate:
		typ = "DATE"
	case sb.ColumnTypeFloat:
		typ = "REAL"
	case sb.ColumnTypeBool:
		typ = "BOOLEAN"
	case sb.ColumnTypeBytes:
		typ = "BLOB"
	}
	if typ == "" {
		return "", errors.New("dialects: unknown column type")
	} else {
		return typ, nil
	}
}

func (m Sqlite) ColumnOptionToString(co *sb.ColumnOption) (string, error) {
	opt := ""
	if co.PrimaryKey {
		opt = str_append(opt, "PRIMARY KEY")
	}
	if co.AutoIncrement {
		opt = str_append(opt, "AUTOINCREMENT")
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
		opt = str_append(opt, "DEFAULT "+m.QuoteField(co.Default))
	}

	return opt, nil
}

func (m Sqlite) TableOptionToString(to *sb.TableOption) (string, error) {
	opt := ""
	if to.Unique != nil {
		opt = str_append(opt, m.tableOptionUnique(to.Unique))
	}

	return "", nil
}

func (m Sqlite) tableOptionUnique(op [][]string) string {
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
