package mymodel_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql"
	"github.com/acoshift/mysql/myctx"
	"github.com/acoshift/mysql/mymodel"
	"github.com/acoshift/mysql/mystmt"
)

func TestDo_SelectModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = myctx.NewContext(ctx, db)

	_, err := db.Exec(`drop table if exists test_mymodel_select;`)
	assert.NoError(t, err)
	_, err = db.Exec(`
		create table test_mymodel_select (
			id int primary key,
			value varchar(255) not null,
			created_at timestamp not null default now()
		);
	`)
	assert.NoError(t, err)
	_, err = db.Exec(`
		insert into test_mymodel_select (id, value)
		values (1, 'value 1'),
			   (2, 'value 2');
	`)
	assert.NoError(t, err)

	{
		var m selectModel
		err = mymodel.Do(ctx, &m, mymodel.Equal("id", 2))
		assert.NoError(t, err)
		assert.Equal(t, int64(2), m.ID)
		assert.Equal(t, "value 2", m.Value)
		assert.NotEmpty(t, m.CreatedAt)
	}

	{
		var m selectModel
		err = mymodel.Do(ctx, &m, mymodel.Equal("id", 99))
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Empty(t, m)
	}

	{
		var ms []*selectModel
		err = mymodel.Do(ctx, &ms, mymodel.OrderBy("id desc"), mymodel.Limit(2))
		assert.NoError(t, err)
		if assert.Len(t, ms, 2) {
			assert.Equal(t, int64(2), ms[0].ID)
			assert.Equal(t, int64(1), ms[1].ID)
		}
	}

	{
		var m selectModel
		err = mymodel.Do(ctx, &m, filterError{})
		assert.Error(t, err)
		_, ok := err.(filterError)
		assert.True(t, ok)
	}
}

type selectModel struct {
	ID        int64
	Value     string
	CreatedAt time.Time
}

func (m *selectModel) Select(b mystmt.SelectStatement) {
	b.Columns("id", "value", "created_at")
	b.From("test_mymodel_select")
}

func (m *selectModel) Scan(scan mysql.Scanner) error {
	return scan(&m.ID, &m.Value, &m.CreatedAt)
}

func TestDo_UpdateModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = myctx.NewContext(ctx, db)

	_, err := db.Exec(`drop table if exists test_mymodel_update;`)
	assert.NoError(t, err)
	_, err = db.Exec(`
		create table test_mymodel_update (
			id int primary key,
			value varchar(255) not null,
			created_at timestamp not null default now(),
			updated_at timestamp null
		);
	`)
	assert.NoError(t, err)
	_, err = db.Exec(`
		insert into test_mymodel_update (id, value)
		values (1, 'value 1'),
			   (2, 'value 2');
	`)
	assert.NoError(t, err)

	{
		err = mymodel.Do(ctx, &updateModel{Value: "new value"}, mymodel.Equal("id", 1))
		assert.NoError(t, err)

		var m updateSelectModel
		err = mymodel.Do(ctx, &m, mymodel.Equal("id", 1))
		assert.NoError(t, err)
		assert.Equal(t, int64(1), m.ID)
		assert.Equal(t, "new value", m.Value)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	}
}

type updateModel struct {
	Value string
}

func (m *updateModel) Update(b mystmt.UpdateStatement) {
	b.Table("test_mymodel_update")
	b.Set("value").To(m.Value)
	b.Set("updated_at").ToRaw("now()")
}

type updateSelectModel struct {
	ID        int64
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *updateSelectModel) Select(b mystmt.SelectStatement) {
	b.Columns("id", "value", "created_at", "updated_at")
	b.From("test_mymodel_update")
}

func (m *updateSelectModel) Scan(scan mysql.Scanner) error {
	return scan(&m.ID, &m.Value, &m.CreatedAt, mysql.NullTime(&m.UpdatedAt))
}

func TestDo_InsertModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = myctx.NewContext(ctx, db)

	_, err := db.Exec(`drop table if exists test_mymodel_insert;`)
	assert.NoError(t, err)
	_, err = db.Exec(`
		create table test_mymodel_insert (
			id int primary key,
			value varchar(255) not null,
			created_at timestamp not null default now()
		);
	`)
	assert.NoError(t, err)

	err = mymodel.Do(ctx, &insertModel{ID: 1, Value: "value 1"})
	assert.NoError(t, err)
}

type insertModel struct {
	ID    int64
	Value string
}

func (m *insertModel) Insert(b mystmt.InsertStatement) {
	b.Into("test_mymodel_insert")
	b.Columns("id", "value")
	b.Value(m.ID, m.Value)
}

type filterError struct{}

func (err filterError) Apply(ctx context.Context, b mymodel.Cond) error {
	return err
}

func (err filterError) Error() string {
	return "error"
}
