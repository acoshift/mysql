package mystmt

// Insert builds insert statement
func Insert(f func(b InsertStatement)) *Result {
	var st insertStmt
	f(&st)
	return newResult(build(st.make()))
}

// InsertStatement is the insert statement builder
type InsertStatement interface {
	Into(table string)
	Columns(col ...string)
	OverridingSystemValue()
	OverridingUserValue()
	DefaultValues()
	Value(value ...interface{})
	Values(values ...interface{})
	Select(f func(b SelectStatement))
	OnDuplicateKey() OnDuplicateKey
}

type OnDuplicateKey interface {
	Update(f func(b UpdateStatement))
}

type insertStmt struct {
	table           string
	columns         parenGroup
	overridingValue string
	defaultValues   bool
	duplicate       *duplicate
	values          group
	selects         *selectStmt
}

func (st *insertStmt) Into(table string) {
	st.table = table
}

func (st *insertStmt) Columns(col ...string) {
	st.columns.pushString(col...)
}

func (st *insertStmt) OverridingSystemValue() {
	st.overridingValue = "system"
}

func (st *insertStmt) OverridingUserValue() {
	st.overridingValue = "user"
}

func (st *insertStmt) DefaultValues() {
	st.defaultValues = true
}

func (st *insertStmt) Value(value ...interface{}) {
	var x parenGroup
	for _, v := range value {
		x.push(Arg(v))
	}
	st.values.push(&x)
}

func (st *insertStmt) Values(values ...interface{}) {
	for _, value := range values {
		st.Value(value)
	}
}

func (st *insertStmt) Select(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)
	st.selects = &x
}

func (st *insertStmt) OnDuplicateKey() OnDuplicateKey {
	st.duplicate = &duplicate{}
	return st.duplicate
}

func (st *insertStmt) make() *buffer {
	var b buffer
	b.push("insert")
	if st.table != "" {
		b.push("into", st.table)
	}
	if !st.columns.empty() {
		b.push(&st.columns)
	}
	if st.overridingValue != "" {
		b.push("overriding", st.overridingValue, "value")
	}
	if st.defaultValues {
		b.push("default values")
	}
	if !st.values.empty() {
		b.push("values", &st.values)
	}
	if st.selects != nil {
		b.push(st.selects.make())
	}
	if st.duplicate != nil {
		b.push("on duplicate key")

		if st.duplicate.update != nil {
			b.push(st.duplicate.update.make())
		}
	}

	return &b
}

type duplicate struct {
	update *updateStmt
}

func (st *duplicate) Update(f func(b UpdateStatement)) {
	var x updateStmt
	f(&x)
	st.update = &x
}
