package mystmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql/mystmt"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		result *mystmt.Result
		query  string
		args   []interface{}
	}{
		{
			"only select",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("1")
			}),
			"select 1",
			nil,
		},
		{
			"select arg",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns(mystmt.Arg("x"))
			}),
			"select ?",
			[]interface{}{
				"x",
			},
		},
		{
			"select without arg",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns(1, "x", 1.2)
			}),
			"select 1, x, 1.2",
			nil,
		},
		{
			"select from",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
			}),
			"select id, name from users",
			nil,
		},
		{
			"select from where",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.Where(func(b mystmt.Cond) {
					b.Eq("id", 3)
					b.Eq("name", "test")
					b.And(func(b mystmt.Cond) {
						b.Eq("age", 15)
						b.Or(func(b mystmt.Cond) {
							b.Eq("age", 18)
						})
					})
					b.Eq("is_active", true)
				})
			}),
			"select id, name from users where (id = ? and name = ? and is_active = ?) and ((age = ?) or (age = ?))",
			[]interface{}{
				3,
				"test",
				true,
				15,
				18,
			},
		},
		{
			"select from where order",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.Where(func(b mystmt.Cond) {
					b.Eq("id", 1)
				})
				b.OrderBy("created_at").Asc().NullsLast()
				b.OrderBy("id").Desc()
			}),
			"select id, name from users where (id = ?) order by created_at asc nulls last, id desc",
			[]interface{}{
				1,
			},
		},
		{
			"select limit offset",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.Where(func(b mystmt.Cond) {
					b.Eq("id", 1)
				})
				b.OrderBy("id")
				b.Limit(5)
				b.Offset(10)
			}),
			"select id, name from users where (id = ?) order by id limit 5 offset 10",
			[]interface{}{
				1,
			},
		},
		{
			"join",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.LeftJoin("roles using id")
			}),
			"select id, name from users left join roles using id",
			nil,
		},
		{
			"join on",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.LeftJoin("roles").On(func(b mystmt.Cond) {
					b.EqRaw("users.id", "roles.id")
				})
			}),
			"select id, name from users left join roles on (users.id = roles.id)",
			nil,
		},
		{
			"join using",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.InnerJoin("roles").Using("id", "name")
			}),
			"select id, name from users inner join roles using (id, name)",
			nil,
		},
		{
			"join select",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("id", "name", "count(*)")
				b.From("users")
				b.LeftJoinSelect(func(b mystmt.SelectStatement) {
					b.Columns("user_id", "data")
					b.From("event")
				}, "t").On(func(b mystmt.Cond) {
					b.EqRaw("t.user_id", "users.id")
				})
				b.GroupBy("id", "name")
			}),
			"select id, name, count(*) from users left join (select user_id, data from event) t on (t.user_id = users.id) group by (id, name)",
			nil,
		},
		{
			"group by having",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("city", "max(temp_lo)")
				b.From("weather")
				b.GroupBy("city")
				b.Having(func(b mystmt.Cond) {
					b.LtRaw("max(temp_lo)", 40)
				})
			}),
			"select city, max(temp_lo) from weather group by (city) having (max(temp_lo) < 40)",
			nil,
		},
		{
			"select in",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.In("x", 1, 2)
				})
			}),
			"select * from table where (x in (?, ?))",
			[]interface{}{
				1,
				2,
			},
		},
		{
			"select in select",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.InSelect("id", func(b mystmt.SelectStatement) {
						b.Columns("id")
						b.From("table2")
					})
				})
			}),
			"select * from table where (id in (select id from table2))",
			nil,
		},
		{
			"select not in",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.NotIn("x", 1, 2)
				})
			}),
			"select * from table where (x not in (?, ?))",
			[]interface{}{
				1,
				2,
			},
		},
		{
			"select and mode",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.Mode().And()
					b.EqRaw("a", 1)
					b.EqRaw("a", 2)
				})
			}),
			"select * from table where (a = 1 and a = 2)",
			nil,
		},
		{
			"select or mode",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.Mode().Or()
					b.EqRaw("a", 1)
					b.EqRaw("a", 2)
				})
			}),
			"select * from table where (a = 1 or a = 2)",
			nil,
		},
		{
			"select nested or mode",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.EqRaw("a", 1)
					b.And(func(b mystmt.Cond) {
						b.Mode().Or()
						b.EqRaw("a", 2)
						b.EqRaw("a", 3)
					})
				})
			}),
			"select * from table where (a = 1) and (a = 2 or a = 3)",
			nil,
		},
		{
			"select nested and",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.EqRaw("a", 1)
					b.EqRaw("b", 1)
					b.And(func(b mystmt.Cond) {
						b.And(func(b mystmt.Cond) {
							b.EqRaw("c", 1)
							b.EqRaw("d", 1)
						})
						b.Or(func(b mystmt.Cond) {
							b.EqRaw("e", 1)
							b.EqRaw("f", 1)
						})
					})
				})
			}),
			"select * from table where (a = 1 and b = 1) and ((c = 1 and d = 1) or (e = 1 and f = 1))",
			nil,
		},
		{
			"select nested and single or without ops",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.EqRaw("a", 1)
					b.EqRaw("b", 1)
					b.And(func(b mystmt.Cond) {
						// nothing to `or` with
						b.Or(func(b mystmt.Cond) {
							b.EqRaw("c", 1)
							b.EqRaw("d", 1)
						})
					})
				})
			}),
			"select * from table where (a = 1 and b = 1) and (c = 1 and d = 1)",
			nil,
		},
		{
			"select without op but nested",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b mystmt.Cond) {
					b.And(func(b mystmt.Cond) {
						b.Mode().Or()
						b.EqRaw("a", 2)
						b.EqRaw("a", 3)
					})
				})
			}),
			"select * from table where (a = 2 or a = 3)",
			nil,
		},
		{
			"select distinct",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Distinct()
				b.Columns("col_1")
			}),
			"select distinct col_1",
			nil,
		},
		{
			"select distinct on",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Distinct().On("col_1", "col_2")
				b.Columns("col_1", "col_3")
			}),
			"select distinct on (col_1, col_2) col_1, col_3",
			nil,
		},
		{
			"left join lateral",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("m.name")
				b.From("manufacturers m")
				b.LeftJoin("lateral get_product_names(m.id) pname").On(func(b mystmt.Cond) {
					b.Raw("true")
				})
				b.Where(func(b mystmt.Cond) {
					b.IsNull("pname")
				})
			}),
			`
				select m.name
				from manufacturers m left join lateral get_product_names(m.id) pname on (true)
				where (pname is null)
			`,
			nil,
		},
		{
			"left join lateral select",
			mystmt.Select(func(b mystmt.SelectStatement) {
				b.Columns("m.name")
				b.From("manufacturers m")
				b.LeftJoinLateralSelect(func(b mystmt.SelectStatement) {
					b.Columns("get_product_names(m.id) pname")
				}, "t").On(func(b mystmt.Cond) {
					b.Raw("true")
				})
				b.Where(func(b mystmt.Cond) {
					b.IsNull("pname")
				})
			}),
			`
				select m.name
				from manufacturers m left join lateral (select get_product_names(m.id) pname) t on (true)
				where (pname is null)
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
