package mymodel

import (
	"context"
	"fmt"
	"reflect"

	"github.com/acoshift/mysql"
	"github.com/acoshift/mysql/mystmt"
)

func Do(ctx context.Context, model interface{}, filter ...Filter) error {
	var err error
	switch m := model.(type) {
	case Selector:
		stmt := mystmt.Select(func(b mystmt.SelectStatement) {
			m.Select(b)
			for _, f := range filter {
				err = f.Apply(ctx, b)
				if err != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}
		return m.Scan(stmt.QueryRowWith(ctx).Scan)
	case Inserter:
		stmt := mystmt.Insert(func(b mystmt.InsertStatement) {
			m.Insert(b)
		})

		if scanner, ok := m.(Scanner); ok {
			return scanner.Scan(stmt.QueryRowWith(ctx).Scan)
		}
		_, err := stmt.ExecWith(ctx)
		return err
	case Updater:
		stmt := mystmt.Update(func(b mystmt.UpdateStatement) {
			m.Update(b)
			for _, f := range filter {
				err = f.Apply(ctx, condUpdateWrapper{b})
				if err != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}

		if scanner, ok := m.(Scanner); ok {
			return scanner.Scan(stmt.QueryRowWith(ctx).Scan)
		}
		_, err := stmt.ExecWith(ctx)
		return err
	}

	// *[]*model => []*model => *model => model
	rf := reflect.ValueOf(model).Elem()
	typeSlice := rf.Type()
	typeElem := typeSlice.Elem().Elem()
	rs := reflect.MakeSlice(typeSlice, 0, 0)
	m := reflect.New(typeElem).Interface()

	if m, ok := m.(Selector); ok {
		stmt := mystmt.Select(func(b mystmt.SelectStatement) {
			m.Select(b)
			for _, f := range filter {
				err = f.Apply(ctx, b)
				if err != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}

		err = stmt.IterWith(ctx, func(scan mysql.Scanner) error {
			rx := reflect.New(typeElem)
			err := rx.Interface().(Selector).Scan(scan)
			if err != nil {
				return err
			}
			rs = reflect.Append(rs, rx)
			return nil
		})
		if err != nil {
			return err
		}
		rf.Set(rs)
		return nil
	}

	return fmt.Errorf("not implement")
}
