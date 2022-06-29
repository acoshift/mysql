package mystmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql/mystmt"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	t.Run("insert", func(t *testing.T) {
		q, args := mystmt.Insert(func(b mystmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "name", "created_at")
			b.Value("tester1", "Tester 1", mystmt.Default)
			b.Value("tester2", "Tester 2", "now()")
			// b.OnDuplicateKey().Update(func(b mystmt.UpdateStatement) {
			// })
		}).SQL()

		assert.Equal(t,
			"insert into users (username, name, created_at) values (?, ?, default), (?, ?, ?)",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				"tester1", "Tester 1",
				"tester2", "Tester 2", "now()",
			},
			args,
		)
	})

	t.Run("insert select", func(t *testing.T) {
		q, args := mystmt.Insert(func(b mystmt.InsertStatement) {
			b.Into("films")
			b.Select(func(b mystmt.SelectStatement) {
				b.Columns("*")
				b.From("tmp_films")
				b.Where(func(b mystmt.Cond) {
					b.LtRaw("date_prod", "2004-05-07")
				})
			})
		}).SQL()

		assert.Equal(t,
			"insert into films select * from tmp_films where (date_prod < 2004-05-07)",
			q,
		)
		assert.Empty(t, args)
	})

	// t.Run("insert on conflict do update", func(t *testing.T) {
	// 	q, args := mystmt.Insert(func(b mystmt.InsertStatement) {
	// 		b.Into("users")
	// 		b.Columns("username", "email")
	// 		b.Value("tester1", "tester1@localhost")
	// 		b.OnDuplicateKey("username").Update(func(b mystmt.UpdateStatement) {
	// 			b.Set("email").ToRaw("excluded.email")
	// 			b.Set("updated_at").ToRaw("now()")
	// 		})
	// 	}).SQL()
	//
	// 	assert.Equal(t,
	// 		stripSpace(`
	// 			insert into users (username, email)
	// 			values (?, ?)
	// 			on conflict (username) do update
	// 			set email = excluded.email,
	// 				updated_at = now()
	// 		`),
	// 		q,
	// 	)
	// 	assert.EqualValues(t,
	// 		[]interface{}{
	// 			"tester1", "tester1@localhost",
	// 		},
	// 		args,
	// 	)
	// })
}
