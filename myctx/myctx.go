package myctx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/acoshift/mysql"
)

type DB interface {
	Queryer
	mysql.BeginTxer
}

// Queryer interface
type Queryer interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// NewContext creates new context
func NewContext(ctx context.Context, db DB) context.Context {
	ctx = context.WithValue(ctx, ctxKeyDB{}, db)
	ctx = context.WithValue(ctx, ctxKeyQueryer{}, db)
	return ctx
}

// Middleware injects db into request's context
func Middleware(db DB) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(NewContext(r.Context(), db))
			h.ServeHTTP(w, r)
		})
	}
}

type wrapTx struct {
	*sql.Tx
	onCommitted []func(ctx context.Context)
}

var _ Queryer = &wrapTx{}

// RunInTxOptions starts sql tx if not started
func RunInTxOptions(ctx context.Context, opt *mysql.TxOptions, f func(ctx context.Context) error) error {
	if IsInTx(ctx) {
		return f(ctx)
	}

	db := ctx.Value(ctxKeyDB{}).(mysql.BeginTxer)
	var pTx wrapTx
	abort := false
	err := mysql.RunInTxContext(ctx, db, opt, func(tx *sql.Tx) error {
		pTx = wrapTx{Tx: tx}
		ctx := context.WithValue(ctx, ctxKeyQueryer{}, &pTx)
		err := f(ctx)
		if errors.Is(err, mysql.ErrAbortTx) {
			abort = true
		}
		return err
	})
	if err != nil {
		return err
	}
	if !abort && len(pTx.onCommitted) > 0 {
		for _, f := range pTx.onCommitted {
			f(ctx)
		}
	}
	return nil
}

// RunInTx calls RunInTxOptions with default options
func RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	return RunInTxOptions(ctx, nil, f)
}

// IsInTx checks is context inside RunInTx
func IsInTx(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyQueryer{}).(*wrapTx)
	return ok
}

// Committed calls f after committed or immediate if not in tx
func Committed(ctx context.Context, f func(ctx context.Context)) {
	if f == nil {
		return
	}

	if !IsInTx(ctx) {
		f(ctx)
		return
	}

	pTx := ctx.Value(ctxKeyQueryer{}).(*wrapTx)
	pTx.onCommitted = append(pTx.onCommitted, f)
}

type (
	ctxKeyDB      struct{}
	ctxKeyQueryer struct{}
)

func q(ctx context.Context) Queryer {
	return ctx.Value(ctxKeyQueryer{}).(Queryer)
}

// QueryRow calls db.QueryRowContext
func QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return q(ctx).QueryRowContext(ctx, query, args...)
}

// Query calls db.QueryContext
func Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return q(ctx).QueryContext(ctx, query, args...)
}

// Exec calls db.ExecContext
func Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q(ctx).ExecContext(ctx, query, args...)
}

// Iter calls mysql.IterContext
func Iter(ctx context.Context, iter mysql.Iterator, query string, args ...interface{}) error {
	return mysql.IterContext(ctx, q(ctx), iter, query, args...)
}

// Prepare calls db.PrepareContext
func Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return q(ctx).PrepareContext(ctx, query)
}
