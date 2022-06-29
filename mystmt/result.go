package mystmt

import (
	"context"
	"database/sql"

	"github.com/acoshift/mysql"
	"github.com/acoshift/mysql/myctx"
)

type Result struct {
	query string
	args  []interface{}
}

func newResult(query string, args []interface{}) *Result {
	return &Result{query, args}
}

func (r *Result) SQL() (query string, args interface{}) {
	return r.query, r.args
}

func (r *Result) QueryRow(f func(string, ...interface{}) *sql.Row) *sql.Row {
	return f(r.query, r.args...)
}

func (r *Result) Query(f func(string, ...interface{}) (*sql.Rows, error)) (*sql.Rows, error) {
	return f(r.query, r.args...)
}

func (r *Result) Exec(f func(string, ...interface{}) (sql.Result, error)) (sql.Result, error) {
	return f(r.query, r.args...)
}

func (r *Result) QueryRowContext(ctx context.Context, f func(context.Context, string, ...interface{}) *sql.Row) *sql.Row {
	return f(ctx, r.query, r.args...)
}

func (r *Result) QueryContext(ctx context.Context, f func(context.Context, string, ...interface{}) (*sql.Rows, error)) (*sql.Rows, error) {
	return f(ctx, r.query, r.args...)
}

func (r *Result) ExecContext(ctx context.Context, f func(context.Context, string, ...interface{}) (sql.Result, error)) (sql.Result, error) {
	return f(ctx, r.query, r.args...)
}

func (r *Result) QueryRowWith(ctx context.Context) *sql.Row {
	return myctx.QueryRow(ctx, r.query, r.args...)
}

func (r *Result) QueryWith(ctx context.Context) (*sql.Rows, error) {
	return myctx.Query(ctx, r.query, r.args...)
}

func (r *Result) ExecWith(ctx context.Context) (sql.Result, error) {
	return myctx.Exec(ctx, r.query, r.args...)
}

func (r *Result) IterWith(ctx context.Context, iter mysql.Iterator) error {
	return myctx.Iter(ctx, iter, r.query, r.args...)
}
