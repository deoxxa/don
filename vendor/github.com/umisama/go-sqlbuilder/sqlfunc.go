package sqlbuilder

// SqlFunc represents function on SQL(ex:count(*)).  This can be use in the same way as Column.
type SqlFunc interface {
	Column

	columns() []Column
}

type sqlFuncColumnList []Column

func (l sqlFuncColumnList) serialize(bldr *builder) {
	first := true
	for _, part := range l {
		if first {
			first = false
		} else {
			bldr.Append(" , ")
		}
		part.serialize(bldr)
	}
}

type sqlFuncImpl struct {
	name string
	args sqlFuncColumnList
}

// Func returns new SQL function.  The name is function name, and the args is arguments of function
func Func(name string, args ...Column) SqlFunc {
	return &sqlFuncImpl{
		name: name,
		args: args,
	}
}

func (m *sqlFuncImpl) As(alias string) Column {
	return &aliasColumn{
		column: m,
		alias:  alias,
	}
}

func (m *sqlFuncImpl) column_name() string {
	return m.name
}

func (m *sqlFuncImpl) not_null() bool {
	return true
}

func (m *sqlFuncImpl) config() ColumnConfig {
	return nil
}

func (m *sqlFuncImpl) acceptType(interface{}) bool {
	return false
}

func (m *sqlFuncImpl) serialize(bldr *builder) {
	bldr.Append(m.name)
	bldr.Append("(")
	bldr.AppendItem(m.args)
	bldr.Append(")")
}

func (left *sqlFuncImpl) Eq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "=")
}

func (left *sqlFuncImpl) NotEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<>")
}

func (left *sqlFuncImpl) Gt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">")
}

func (left *sqlFuncImpl) GtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, ">=")
}

func (left *sqlFuncImpl) Lt(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<")
}

func (left *sqlFuncImpl) LtEq(right interface{}) Condition {
	return newBinaryOperationCondition(left, right, "<=")
}

func (left *sqlFuncImpl) Like(right string) Condition {
	return newBinaryOperationCondition(left, right, " LIKE ")
}

func (left *sqlFuncImpl) Between(lower, higher interface{}) Condition {
	return newBetweenCondition(left, lower, higher)
}

func (left *sqlFuncImpl) In(vals ...interface{}) Condition {
	return newInCondition(left, vals...)
}

func (left *sqlFuncImpl) IntersectJSON(data string) Condition {
	return newIntersectJSONCondition(left, data)
}

func (left *sqlFuncImpl) NotIn(val ...interface{}) Condition {
	return newNotInCondition(left, val...)
}

func (m *sqlFuncImpl) columns() []Column {
	return m.args
}
