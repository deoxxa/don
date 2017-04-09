package dialects

import (
	"errors"
	"fmt"
	sb "github.com/umisama/go-sqlbuilder"
	"time"
)

type MySql struct{}

func (m MySql) QuerySuffix() string {
	return ";"
}

func (m MySql) BindVar(i int) string {
	return "?"
}

func (m MySql) QuoteField(field interface{}) string {
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
		str = "`" + str + "`"
	}
	return str
}

func (m MySql) ColumnTypeToString(cc sb.ColumnConfig) (string, error) {
	if cc.Option().SqlType != "" {
		return cc.Option().SqlType, nil
	}

	typ := ""
	switch cc.Type() {
	case sb.ColumnTypeInt:
		typ = "INTEGER"
	case sb.ColumnTypeString:
		typ = fmt.Sprintf("VARCHAR(%d)", cc.Option().Size)
	case sb.ColumnTypeDate:
		typ = "DATETIME"
	case sb.ColumnTypeFloat:
		typ = "FLOAT"
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

func (m MySql) ColumnOptionToString(co *sb.ColumnOption) (string, error) {
	opt := ""
	if co.PrimaryKey {
		opt = str_append(opt, "PRIMARY KEY")
	}
	if co.AutoIncrement {
		opt = str_append(opt, "AUTO_INCREMENT")
	}
	if co.NotNull {
		opt = str_append(opt, "NOT NULL")
	}
	if co.Unique {
		opt = str_append(opt, "UNIQUE")
	}

	return opt, nil
}

func (m MySql) TableOptionToString(to *sb.TableOption) (string, error) {
	opt := ""
	if to.Unique != nil {
		opt = str_append(opt, m.tableOptionUnique(to.Unique))
	}

	return "", nil
}

func (m MySql) tableOptionUnique(op [][]string) string {
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

func str_append(str, opt string) string {
	if len(str) != 0 {
		str += " "
	}
	str += opt
	return str
}
