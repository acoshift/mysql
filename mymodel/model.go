package mymodel

import (
	"github.com/acoshift/mysql"
	"github.com/acoshift/mysql/mystmt"
)

// Scanner model
type Scanner interface {
	Scan(scan mysql.Scanner) error
}

// Selector model
type Selector interface {
	Select(b mystmt.SelectStatement)
	Scanner
}

// Inserter model
type Inserter interface {
	Insert(b mystmt.InsertStatement)
}

// Updater model
type Updater interface {
	Update(b mystmt.UpdateStatement)
}
