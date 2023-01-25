package mystmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql/mystmt"
)

func TestUnion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		result *mystmt.Result
		query  string
		args   []interface{}
	}{
		{
			"union select",
			mystmt.Union(func(b mystmt.UnionStatement) {
				b.Select(func(b mystmt.SelectStatement) {
					b.Columns("id")
					b.From("table1")
				})
				b.AllSelect(func(b mystmt.SelectStatement) {
					b.Columns("id")
					b.From("table2")
				})
				b.DistinctSelect(func(b mystmt.SelectStatement) {
					b.Columns("id")
					b.From("table3")
				})
				b.OrderBy("id")
				b.Limit(10)
			}),
			`
				(select id from table1)
				union all (select id from table2)
				union distinct (select id from table3)
				order by id
				limit 10
			`,
			nil,
		},
	}

	for _, tC := range cases {
		t.Run(tC.name, func(t *testing.T) {
			q, args := tC.result.SQL()
			assert.Equal(t, stripSpace(tC.query), q)
			assert.EqualValues(t, tC.args, args)
		})
	}
}
