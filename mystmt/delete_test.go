package mystmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql/mystmt"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	q, args := mystmt.Delete(func(b mystmt.DeleteStatement) {
		b.From("users")
		b.Where(func(b mystmt.Cond) {
			b.Eq("username", "test")
			b.Eq("is_active", false)
			b.Or(func(b mystmt.Cond) {
				b.Gt("age", mystmt.Arg(20))
				b.Le("age", mystmt.Arg(30))
			})
		})
	}).SQL()

	assert.Equal(t,
		"delete from users where (username = ? and is_active = ?) or (age > ? and age <= ?)",
		q,
	)
	assert.EqualValues(t,
		[]interface{}{"test", false, 20, 30},
		args,
	)
}
