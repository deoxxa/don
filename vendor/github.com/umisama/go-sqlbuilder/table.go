package sqlbuilder

type joinType int

const (
	inner_join joinType = iota
	left_outer_join
	right_outer_join
	full_outer_join
)

type table struct {
	name    string
	option  *TableOption
	columns []Column
	alias   string
}

// TableOption reprecents constraint of a table.
type TableOption struct {
	Unique [][]string
	//ForeignKey map[string]Column // will implement future
}

type joinTable struct {
	typ   joinType
	left  Table
	right Table
	on    Condition
	alias string
}

// Table represents a table.
type Table interface {
	serializable

	// As returns a copy of the table with an alias.
	As(alias string) Table

	// C returns table's column by the name.
	C(name string) Column

	// Name returns table' name.
	// returns empty if it is joined table or subquery.
	Name() string

	// Option returns table's option(table constraint).
	// returns nil if it is joined table or subquery.
	Option() *TableOption

	// Columns returns all columns.
	Columns() []Column

	// InnerJoin returns a joined table use with "INNER JOIN" clause.
	// The joined table can be handled in same way as single table.
	InnerJoin(Table, Condition) Table

	// LeftOuterJoin returns a joined table use with "LEFT OUTER JOIN" clause.
	// The joined table can be handled in same way as single table.
	LeftOuterJoin(Table, Condition) Table

	// RightOuterJoin returns a joined table use with "RIGHT OUTER JOIN" clause.
	// The joined table can be handled in same way as single table.
	RightOuterJoin(Table, Condition) Table

	// FullOuterJoin returns a joined table use with "FULL OUTER JOIN" clause.
	// The joined table can be handled in same way as single table.
	FullOuterJoin(Table, Condition) Table

	hasColumn(column Column) bool
}

// NewTable returns a new table named by the name.  Specify table columns by the column_config.
// Panic if column is empty.
func NewTable(name string, option *TableOption, column_configs ...ColumnConfig) Table {
	if len(column_configs) == 0 {
		panic(newError("column is needed."))
	}
	if option == nil {
		option = &TableOption{}
	}

	t := &table{
		name:    name,
		option:  option,
		columns: make([]Column, 0, len(column_configs)),
	}

	for _, column_config := range column_configs {
		err := t.AddColumnLast(column_config)
		if err != nil {
			panic(err)
		}
	}

	return t
}

func (m *table) serialize(bldr *builder) {
	bldr.Append(dialect().QuoteField(m.name))
	if m.alias != "" {
		bldr.Append(" " + m.alias)
	}
	return
}

func (m *table) As(alias string) Table {
	t := &table{
		name:    m.name,
		option:  m.option,
		alias:   alias,
		columns: make([]Column, len(m.columns)),
	}

	for i, c := range m.columns {
		if cc, ok := c.(*columnImpl); ok {
			c = &columnImpl{
				columnConfigImpl: cc.columnConfigImpl,
				table:            t,
			}
		}

		t.columns[i] = c
	}

	return t
}

func (m *table) C(name string) Column {
	for _, column := range m.columns {
		if column.column_name() == name {
			return column
		}
	}

	return newErrorColumn(newError("column %s.%s was not found.", m.name, name))
}

func (m *table) Name() string {
	if m.alias != "" {
		return m.alias
	}

	return m.name
}

func (m *table) SetName(name string) {
	m.name = name
}

func (m *table) Columns() []Column {
	return m.columns
}

func (m *table) Option() *TableOption {
	return m.option
}

func (m *table) AddColumnLast(cc ColumnConfig) error {
	return m.addColumn(cc, len(m.columns))
}

func (m *table) AddColumnFirst(cc ColumnConfig) error {
	return m.addColumn(cc, 0)
}

func (m *table) AddColumnAfter(cc ColumnConfig, after Column) error {
	for i := range m.columns {
		if m.columns[i] == after {
			return m.addColumn(cc, i+1)
		}
	}
	return newError("column not found.")
}

func (m *table) ChangeColumn(trg Column, cc ColumnConfig) error {
	for i := range m.columns {
		if m.columns[i] == trg {
			err := m.dropColumn(i)
			if err != nil {
				return err
			}
			err = m.addColumn(cc, i)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return newError("column not found.")
}

func (m *table) ChangeColumnFirst(trg Column, cc ColumnConfig) error {
	for i := range m.columns {
		if m.columns[i] == trg {
			err := m.dropColumn(i)
			if err != nil {
				return err
			}
			err = m.addColumn(cc, 0)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return newError("column not found.")
}

func (m *table) ChangeColumnAfter(trg Column, cc ColumnConfig, after Column) error {
	backup := make([]Column, len(m.columns))
	copy(backup, m.columns)
	found := false
	for i := range m.columns {
		if m.columns[i] == trg {
			err := m.dropColumn(i)
			if err != nil {
				m.columns = backup
				return err
			}
			found = true
			break
		}
	}
	if !found {
		return newError("column not found.")
	}
	for i := range m.columns {
		if m.columns[i] == after {
			err := m.addColumn(cc, i+1)
			if err != nil {
				m.columns = backup
				return err
			}
			return nil
		}
	}
	m.columns = backup
	return newError("column not found.")
}

func (m *table) addColumn(cc ColumnConfig, pos int) error {
	if len(m.columns) < pos || pos < 0 {
		return newError("Invalid position.")
	}

	var (
		u = make([]Column, pos)
		p = make([]Column, len(m.columns)-pos)
	)
	copy(u, m.columns[:pos])
	copy(p, m.columns[pos:])
	c := cc.toColumn(m)
	m.columns = append(u, c)
	m.columns = append(m.columns, p...)
	return nil
}

func (m *table) DropColumn(col Column) error {
	for i := range m.columns {
		if m.columns[i] == col {
			return m.dropColumn(i)
		}
	}
	return newError("column not found.")
}

func (m *table) dropColumn(pos int) error {
	if len(m.columns) < pos || pos < 0 {
		return newError("Invalid position.")
	}
	var (
		u = make([]Column, pos)
		p = make([]Column, len(m.columns)-pos-1)
	)
	copy(u, m.columns[:pos])
	if len(m.columns) > pos+1 {
		copy(p, m.columns[pos+1:])
	}
	m.columns = append(u, p...)
	return nil
}

func (m *table) InnerJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   inner_join,
		on:    on,
	}
}

func (m *table) LeftOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   left_outer_join,
		on:    on,
	}
}

func (m *table) RightOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   right_outer_join,
		on:    on,
	}
}

func (m *table) FullOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   full_outer_join,
		on:    on,
	}
}

func (m *table) hasColumn(trg Column) bool {
	if cimpl, ok := trg.(*columnImpl); ok {
		if trg == Star {
			return true
		}
		for _, col := range m.columns {
			if col == cimpl {
				return true
			}
		}
		return false
	}
	if acol, ok := trg.(*aliasColumn); ok {
		for _, col := range m.columns {
			if col == acol.column {
				return true
			}
		}
		return false
	}
	if sqlfn, ok := trg.(*sqlFuncImpl); ok {
		for _, fncol := range sqlfn.columns() {
			find := false
			for _, col := range m.columns {
				if col == fncol {
					find = true
				}
			}
			if !find {
				return false
			}
		}
		return true
	}
	return false
}

func (m *joinTable) As(string) Table {
	return m
}

func (m *joinTable) C(name string) Column {
	l_col := m.left.C(name)
	r_col := m.right.C(name)

	_, l_err := l_col.(*errorColumn)
	_, r_err := r_col.(*errorColumn)

	switch {
	case l_err && r_err:
		return newErrorColumn(newError("column %s was not found.", name))
	case l_err && !r_err:
		return r_col
	case !l_err && r_err:
		return l_col
	default:
		return newErrorColumn(newError("column %s was duplicated.", name))
	}
}

func (m *joinTable) Name() string {
	return ""
}

func (m *joinTable) Columns() []Column {
	return append(m.left.Columns(), m.right.Columns()...)
}

func (m *joinTable) Option() *TableOption {
	return nil
}

func (m *joinTable) InnerJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   inner_join,
		on:    on,
	}
}

func (m *joinTable) LeftOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   left_outer_join,
		on:    on,
	}
}

func (m *joinTable) RightOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   right_outer_join,
		on:    on,
	}
}

func (m *joinTable) FullOuterJoin(right Table, on Condition) Table {
	return &joinTable{
		left:  m,
		right: right,
		typ:   full_outer_join,
		on:    on,
	}
}

func (m *joinTable) serialize(bldr *builder) {
	bldr.AppendItem(m.left)
	switch m.typ {
	case inner_join:
		bldr.Append(" INNER JOIN ")
	case left_outer_join:
		bldr.Append(" LEFT OUTER JOIN ")
	case right_outer_join:
		bldr.Append(" RIGHT OUTER JOIN ")
	case full_outer_join:
		bldr.Append(" FULL OUTER JOIN ")
	}
	bldr.AppendItem(m.right)
	bldr.Append(" ON ")
	bldr.AppendItem(m.on)
	return
}

func (m *joinTable) hasColumn(trg Column) bool {
	if m.left.hasColumn(trg) {
		return true
	}
	if m.right.hasColumn(trg) {
		return true
	}
	return false
}
