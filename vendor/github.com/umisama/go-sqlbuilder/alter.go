package sqlbuilder

type AlterTableStatement struct {
	table          *table
	rename_to      string
	add_columns    []*alterTableAddColumn
	drop_columns   []Column
	change_columns []*alterTableChangeColumn

	err error
}

func AlterTable(tbl Table) *AlterTableStatement {
	if tbl == nil {
		return &AlterTableStatement{
			err: newError("table is nil."),
		}
	}

	t, ok := tbl.(*table)
	if !ok {
		return &AlterTableStatement{
			err: newError("AlterTable can use only natural table."),
		}
	}
	return &AlterTableStatement{
		table:          t,
		add_columns:    make([]*alterTableAddColumn, 0),
		change_columns: make([]*alterTableChangeColumn, 0),
	}
}

func (b *AlterTableStatement) RenameTo(name string) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.rename_to = name
	return b
}

func (b *AlterTableStatement) AddColumn(col ColumnConfig) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.add_columns = append(b.add_columns, &alterTableAddColumn{
		table:  b.table,
		column: col,
		first:  false,
		after:  nil,
	})
	return b
}

func (b *AlterTableStatement) AddColumnAfter(col ColumnConfig, after Column) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.add_columns = append(b.add_columns, &alterTableAddColumn{
		table:  b.table,
		column: col,
		first:  false,
		after:  after,
	})
	return b
}

func (b *AlterTableStatement) AddColumnFirst(col ColumnConfig) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.add_columns = append(b.add_columns, &alterTableAddColumn{
		table:  b.table,
		column: col,
		first:  true,
		after:  nil,
	})
	return b
}

func (b *AlterTableStatement) DropColumn(col Column) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.drop_columns = append(b.drop_columns, col)
	return b
}

func (b *AlterTableStatement) ChangeColumn(old_column Column, new_column ColumnConfig) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.change_columns = append(b.change_columns, &alterTableChangeColumn{
		table:      b.table,
		old_column: old_column,
		new_column: new_column,
		first:      false,
		after:      nil,
	})
	return b
}

func (b *AlterTableStatement) ChangeColumnAfter(old_column Column, new_column ColumnConfig, after Column) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.change_columns = append(b.change_columns, &alterTableChangeColumn{
		table:      b.table,
		old_column: old_column,
		new_column: new_column,
		first:      false,
		after:      after,
	})
	return b
}

func (b *AlterTableStatement) ChangeColumnFirst(old_column Column, new_column ColumnConfig) *AlterTableStatement {
	if b.err != nil {
		return b
	}

	b.change_columns = append(b.change_columns, &alterTableChangeColumn{
		table:      b.table,
		old_column: old_column,
		new_column: new_column,
		first:      true,
		after:      nil,
	})
	return b
}

func (b *AlterTableStatement) ToSql() (query string, args []interface{}, err error) {
	bldr := newBuilder()
	defer func() {
		query, args, err = bldr.Query(), bldr.Args(), bldr.Err()
	}()
	if b.err != nil {
		bldr.SetError(b.err)
		return
	}

	bldr.Append("ALTER TABLE ")
	bldr.AppendItem(b.table)
	bldr.Append(" ")

	first := true
	for _, add_column := range b.add_columns {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		bldr.AppendItem(add_column)
	}
	for _, change_column := range b.change_columns {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		bldr.AppendItem(change_column)
	}
	for _, drop_column := range b.drop_columns {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		bldr.Append("DROP COLUMN ")
		if colname := drop_column.column_name(); len(colname) != 0 {
			bldr.Append(dialect().QuoteField(colname))
		} else {
			bldr.AppendItem(drop_column)
		}
	}
	if len(b.rename_to) != 0 {
		if first {
			first = false
		} else {
			bldr.Append(", ")
		}
		bldr.Append("RENAME TO ")
		bldr.Append(dialect().QuoteField(b.rename_to))
	}

	return "", nil, nil
}

func (b *AlterTableStatement) ApplyToTable() error {
	for _, add_column := range b.add_columns {
		err := add_column.applyToTable()
		if err != nil {
			return err
		}
	}
	for _, change_column := range b.change_columns {
		err := change_column.applyToTable()
		if err != nil {
			return err
		}
	}
	for _, drop_column := range b.drop_columns {
		err := b.table.DropColumn(drop_column)
		if err != nil {
			return err
		}
	}
	if len(b.rename_to) != 0 {
		b.table.SetName(b.rename_to)
	}
	return nil
}

type alterTableAddColumn struct {
	table  *table
	column ColumnConfig
	first  bool
	after  Column
}

func (b *alterTableAddColumn) serialize(bldr *builder) {
	bldr.Append("ADD COLUMN ")
	bldr.AppendItem(b.column)

	// SQL data name
	typ, err := dialect().ColumnTypeToString(b.column)
	if err != nil {
		bldr.SetError(err)
	} else if len(typ) == 0 {
		bldr.SetError(newError("column type is required.(maybe, a bug is in implements of dialect.)"))
	} else {
		bldr.Append(" ")
		bldr.Append(typ)
	}

	opt, err := dialect().ColumnOptionToString(b.column.Option())
	if err != nil {
		bldr.SetError(err)
	} else if len(opt) != 0 {
		bldr.Append(" ")
		bldr.Append(opt)
	}

	if b.first {
		bldr.Append(" FIRST")
	} else if b.after != nil {
		bldr.Append(" AFTER ")
		if colname := b.after.column_name(); len(colname) != 0 {
			bldr.Append(dialect().QuoteField(colname))
		} else {
			bldr.AppendItem(b.after)
		}
	}
}

func (b *alterTableAddColumn) applyToTable() error {
	if b.first {
		return b.table.AddColumnFirst(b.column)
	}
	if b.after != nil {
		return b.table.AddColumnAfter(b.column, b.after)
	}
	return b.table.AddColumnLast(b.column)
}

type alterTableChangeColumn struct {
	table      *table
	old_column Column
	new_column ColumnConfig
	first      bool
	after      Column
}

func (b *alterTableChangeColumn) serialize(bldr *builder) {
	bldr.Append("CHANGE COLUMN ")
	if colname := b.old_column.column_name(); len(colname) != 0 {
		bldr.Append(dialect().QuoteField(colname))
	} else {
		bldr.AppendItem(b.old_column)
	}
	bldr.Append(" ")
	bldr.AppendItem(b.new_column)

	typ, err := dialect().ColumnTypeToString(b.new_column)
	if err != nil {
		bldr.SetError(err)
	} else if len(typ) == 0 {
		bldr.SetError(newError("column type is required.(maybe, a bug is in implements of dialect.)"))
	} else {
		bldr.Append(" ")
		bldr.Append(typ)
	}

	opt, err := dialect().ColumnOptionToString(b.new_column.Option())
	if err != nil {
		bldr.SetError(err)
	} else if len(opt) != 0 {
		bldr.Append(" ")
		bldr.Append(opt)
	}

	if b.first {
		bldr.Append(" FIRST")
	} else if b.after != nil {
		bldr.Append(" AFTER ")
		if colname := b.after.column_name(); len(colname) != 0 {
			bldr.Append(dialect().QuoteField(colname))
		} else {
			bldr.AppendItem(b.after)
		}
	}
}

func (b *alterTableChangeColumn) applyToTable() error {
	if b.first {
		return b.table.ChangeColumnFirst(b.old_column, b.new_column)
	}
	if b.after != nil {
		return b.table.ChangeColumnAfter(b.old_column, b.new_column, b.after)
	}
	return b.table.ChangeColumn(b.old_column, b.new_column)
}
