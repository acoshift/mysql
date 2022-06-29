package mymodel

import (
	"context"

	"github.com/acoshift/mysql/mystmt"
)

type Cond interface {
	Where(f func(b mystmt.Cond))
	Having(f func(b mystmt.Cond))
	OrderBy(col string) mystmt.OrderBy
	Limit(n int64)
	Offset(n int64)
}

type Filter interface {
	Apply(ctx context.Context, b Cond) error
}

type FilterFunc func(ctx context.Context, b Cond) error

func (f FilterFunc) Apply(ctx context.Context, b Cond) error { return f(ctx, b) }

func Equal(field string, value interface{}) Filter {
	return Where(func(b mystmt.Cond) {
		b.Eq(field, value)
	})
}

func Where(f func(b mystmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b Cond) error {
		b.Where(f)
		return nil
	})
}

func Having(f func(b mystmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b Cond) error {
		b.Having(f)
		return nil
	})
}

func OrderBy(col string) Filter {
	return FilterFunc(func(_ context.Context, b Cond) error {
		b.OrderBy(col)
		return nil
	})
}

func Limit(n int64) Filter {
	return FilterFunc(func(_ context.Context, b Cond) error {
		b.Limit(n)
		return nil
	})
}

func Offset(n int64) Filter {
	return FilterFunc(func(_ context.Context, b Cond) error {
		b.Offset(n)
		return nil
	})
}

type condUpdateWrapper struct {
	mystmt.UpdateStatement
}

func (c condUpdateWrapper) Having(f func(b mystmt.Cond)) {}

func (c condUpdateWrapper) OrderBy(col string) mystmt.OrderBy { return noopOrderBy{} }

func (c condUpdateWrapper) Limit(n int64) {}

func (c condUpdateWrapper) Offset(n int64) {}

type noopOrderBy struct{}

func (n noopOrderBy) Asc() mystmt.OrderBy { return n }

func (n noopOrderBy) Desc() mystmt.OrderBy { return n }

func (n noopOrderBy) NullsFirst() mystmt.OrderBy { return n }

func (n noopOrderBy) NullsLast() mystmt.OrderBy { return n }
