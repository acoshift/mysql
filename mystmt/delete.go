package mystmt

// Delete builds delete statement
func Delete(f func(b DeleteStatement)) *Result {
	var st deleteStmt
	f(&st)
	return newResult(build(st.make()))
}

type DeleteStatement interface {
	From(table string)
	Where(f func(b Cond))
}

type deleteStmt struct {
	from  string
	where cond
}

func (st *deleteStmt) From(table string) {
	st.from = table
}

func (st *deleteStmt) Where(f func(b Cond)) {
	f(&st.where)
}

func (st *deleteStmt) make() *buffer {
	var b buffer
	b.push("delete from", st.from)
	if !st.where.empty() {
		b.push("where")
		b.push(st.where.build()...)
	}

	return &b
}
