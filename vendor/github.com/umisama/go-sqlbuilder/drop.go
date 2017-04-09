package sqlbuilder

// DeleteTableStatement represents a "DROP TABLE" statement.
type DropTableStatement struct {
	table Table

	err error
}

// DropTable returns new "DROP TABLE" statement. The table is Table object to drop.
func DropTable(tbl Table) *DropTableStatement {
	if tbl == nil {
		return &DropTableStatement{
			err: newError("table is nil."),
		}
	}
	if _, ok := tbl.(*table); !ok {
		return &DropTableStatement{
			err: newError("table is not natural table."),
		}
	}
	return &DropTableStatement{
		table: tbl,
	}
}

// ToSql generates query string, placeholder arguments, and returns err on errors.
func (b *DropTableStatement) ToSql() (query string, args []interface{}, err error) {
	bldr := newBuilder()
	defer func() {
		query, args, err = bldr.Query(), bldr.Args(), bldr.Err()
	}()
	if b.err != nil {
		bldr.SetError(b.err)
		return
	}

	bldr.Append("DROP TABLE ")
	bldr.AppendItem(b.table)
	return
}
