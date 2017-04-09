package sqlbuilder

// CreateIndexStatement represents a "CREATE INDEX" statement.
type CreateIndexStatement struct {
	table       Table
	columns     []Column
	name        string
	ifNotExists bool

	err error
}

// CreateTableStatement represents a "CREATE TABLE" statement.
type CreateTableStatement struct {
	table       Table
	ifNotExists bool

	err error
}

// CreateTable returns new "CREATE TABLE" statement. The table is Table object to create.
func CreateTable(tbl Table) *CreateTableStatement {
	if tbl == nil {
		return &CreateTableStatement{
			err: newError("table is nil."),
		}
	}
	if _, ok := tbl.(*table); !ok {
		return &CreateTableStatement{
			err: newError("CreateTable can use only natural table."),
		}
	}

	return &CreateTableStatement{
		table: tbl,
	}
}

// IfNotExists sets "IF NOT EXISTS" clause.
func (b *CreateTableStatement) IfNotExists() *CreateTableStatement {
	if b.err != nil {
		return b
	}
	b.ifNotExists = true
	return b
}

// ToSql generates query string, placeholder arguments, and error.
func (b *CreateTableStatement) ToSql() (query string, args []interface{}, err error) {
	bldr := newBuilder()
	defer func() {
		query, args, err = bldr.Query(), bldr.Args(), bldr.Err()
	}()
	if b.err != nil {
		bldr.SetError(b.err)
		return
	}

	bldr.Append("CREATE TABLE ")
	if b.ifNotExists {
		bldr.Append("IF NOT EXISTS ")
	}
	bldr.AppendItem(b.table)

	if len(b.table.Columns()) != 0 {
		bldr.Append(" ( ")
		bldr.AppendItem(createTableColumnList(b.table.Columns()))
		bldr.Append(" )")
	} else {
		bldr.SetError(newError("CreateTableStatement needs one or more columns."))
		return
	}

	// table option
	if tabopt, err := dialect().TableOptionToString(b.table.Option()); err == nil {
		if len(tabopt) != 0 {
			bldr.Append(" " + tabopt)
		}
	} else {
		bldr.SetError(err)
	}

	return
}

// CreateIndex returns new "CREATE INDEX" statement. The table is Table object to create index.
func CreateIndex(tbl Table) *CreateIndexStatement {
	if tbl == nil {
		return &CreateIndexStatement{
			err: newError("table is nil."),
		}
	}
	if _, ok := tbl.(*table); !ok {
		return &CreateIndexStatement{
			err: newError("CreateTable can use only natural table."),
		}
	}
	return &CreateIndexStatement{
		table: tbl,
	}
}

// IfNotExists sets "IF NOT EXISTS" clause.
func (b *CreateIndexStatement) IfNotExists() *CreateIndexStatement {
	if b.err != nil {
		return b
	}
	b.ifNotExists = true
	return b
}

// IfNotExists sets "IF NOT EXISTS" clause. If not set this, returns error on ToSql().
func (b *CreateIndexStatement) Columns(columns ...Column) *CreateIndexStatement {
	if b.err != nil {
		return b
	}
	b.columns = columns
	return b
}

// Name sets name for index.
// If not set this, auto generated name will be used.
func (b *CreateIndexStatement) Name(name string) *CreateIndexStatement {
	if b.err != nil {
		return b
	}
	b.name = name
	return b
}

// ToSql generates query string, placeholder arguments, and returns err on errors.
func (b *CreateIndexStatement) ToSql() (query string, args []interface{}, err error) {
	bldr := newBuilder()
	defer func() {
		query, args, err = bldr.Query(), bldr.Args(), bldr.Err()
	}()
	if b.err != nil {
		bldr.SetError(b.err)
		return
	}

	bldr.Append("CREATE INDEX ")
	if b.ifNotExists {
		bldr.Append("IF NOT EXISTS ")
	}

	if len(b.name) != 0 {
		bldr.Append(dialect().QuoteField(b.name))
	} else {
		bldr.SetError(newError("name was not setted."))
		return
	}

	bldr.Append(" ON ")
	bldr.AppendItem(b.table)

	if len(b.columns) != 0 {
		bldr.Append(" ( ")
		bldr.AppendItem(createIndexColumnList(b.columns))
		bldr.Append(" )")
	} else {
		bldr.SetError(newError("columns was not setted."))
		return
	}
	return
}

type createTableColumnList []Column

func (m createTableColumnList) serialize(bldr *builder) {
	first := true
	for _, column := range m {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		cc := column.config()

		// Column name
		bldr.AppendItem(cc)
		bldr.Append(" ")

		// SQL data name
		str, err := dialect().ColumnTypeToString(cc)
		if err != nil {
			bldr.SetError(err)
		}
		bldr.Append(str)

		str, err = dialect().ColumnOptionToString(cc.Option())
		if err != nil {
			bldr.SetError(err)
		}
		if len(str) != 0 {
			bldr.Append(" " + str)
		}
	}
}

type createIndexColumnList []Column

func (m createIndexColumnList) serialize(bldr *builder) {
	first := true
	for _, column := range m {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		cc := column.config()
		bldr.AppendItem(cc)
	}
}
