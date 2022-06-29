package mystmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql/mystmt"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("update", func(t *testing.T) {
		q, args := mystmt.Update(func(b mystmt.UpdateStatement) {
			b.Table("users")
			b.Set("name").To("test")
			b.Set("email", "address", "updated_at").To("test@localhost", "123", mystmt.NotArg("now()"))
			b.Set("age").ToRaw(1)
			b.Where(func(b mystmt.Cond) {
				b.Eq("id", 5)
			})
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set name = ?,
					(email, address, updated_at) = row(?, ?, now()),
					age = 1
				where (id = ?)
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				"test",
				"test@localhost", "123",
				5,
			},
			args,
		)
	})

	t.Run("update set select", func(t *testing.T) {
		q, args := mystmt.Update(func(b mystmt.UpdateStatement) {
			b.Table("users")
			b.Set("name", "age", "updated_at").Select(func(b mystmt.SelectStatement) {
				b.Columns("name", "age", "now()")
				b.From("users")
				b.Where(func(b mystmt.Cond) {
					b.Eq("id", 6)
				})
			})
			b.Set("updated_count").ToRaw("updated_count + 1")
			b.Set("email", "address").To("test@localhost", "123")
			b.Where(func(b mystmt.Cond) {
				b.Eq("id", 5)
			})
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set (name, age, updated_at) = (select name, age, now()
											   from users
											   where (id = ?)),
					updated_count = updated_count + 1,
					(email, address) = row(?, ?)
				where (id = ?)
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				6,
				"test@localhost", "123",
				5,
			},
			args,
		)
	})

	t.Run("update from join", func(t *testing.T) {
		q, args := mystmt.Update(func(b mystmt.UpdateStatement) {
			b.Table("users")
			b.Set("name").ToRaw("p.name")
			b.Set("address").ToRaw("p.address")
			b.Set("updated_at").ToRaw("now()")
			b.From("users")
			b.InnerJoin("profiles p").Using("email")
			b.Where(func(b mystmt.Cond) {
				b.Eq("users.id", 2)
			})
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set name = p.name,
					address = p.address,
					updated_at = now()
				from users
				inner join profiles p using (email)
				where (users.id = ?)
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				2,
			},
			args,
		)
	})
}
