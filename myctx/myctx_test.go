package myctx_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/acoshift/mysql"
	"github.com/acoshift/mysql/myctx"
)

func newCtx(t *testing.T) (context.Context, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	return myctx.NewContext(context.Background(), db), mock
}

func TestNewContext(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		newCtx(t)
	})
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	called := false
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	myctx.Middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		ctx := r.Context()
		assert.NotPanics(t, func() {
			myctx.QueryRow(ctx, "select 1")
		})
		assert.NotPanics(t, func() {
			myctx.Query(ctx, "select 1")
		})
		assert.NotPanics(t, func() {
			myctx.Exec(ctx, "select 1")
		})
	})).ServeHTTP(w, r)
	assert.True(t, called)
}

func TestRunInTx(t *testing.T) {
	t.Parallel()

	t.Run("Committed", func(t *testing.T) {
		ctx, mock := newCtx(t)

		called := false
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			called = true
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Rollback with error", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectRollback()
		var retErr = fmt.Errorf("error")
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			return retErr
		})
		assert.Error(t, err)
		assert.Equal(t, retErr, err)
	})

	t.Run("Abort Tx", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			return mysql.ErrAbortTx
		})
		assert.NoError(t, err)
	})

	t.Run("Nested Tx", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			return myctx.RunInTx(ctx, func(ctx context.Context) error {
				return nil
			})
		})
		assert.NoError(t, err)
	})
}

func TestCommitted(t *testing.T) {
	t.Parallel()

	t.Run("Outside Tx", func(t *testing.T) {
		ctx, _ := newCtx(t)
		var called bool
		myctx.Committed(ctx, func(ctx context.Context) {
			called = true
		})
		assert.True(t, called)
	})

	t.Run("Nil func", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		myctx.RunInTx(ctx, func(ctx context.Context) error {
			myctx.Committed(ctx, nil)
			return nil
		})
	})

	t.Run("Committed", func(t *testing.T) {
		ctx, mock := newCtx(t)

		called := false
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			myctx.Committed(ctx, func(ctx context.Context) {
				called = true
			})
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Rollback", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectRollback()
		err := myctx.RunInTx(ctx, func(ctx context.Context) error {
			myctx.Committed(ctx, func(ctx context.Context) {
				assert.Fail(t, "should not be called")
			})
			return mysql.ErrAbortTx
		})
		assert.NoError(t, err)
	})
}
