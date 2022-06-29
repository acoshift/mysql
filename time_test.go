package mysql_test

import (
	"testing"
	"time"

	"github.com/acoshift/mysql"
)

func TestTime(t *testing.T) {
	db := open(t)

	_, err := db.Exec(`drop table if exists test_mysql_time;`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	_, err = db.Exec(`
		create table test_mysql_time (
			id int primary key,
			value timestamp null
		);
	`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	_, err = db.Exec(`
		insert into test_mysql_time (
			id, value
		) values
			(0, now()),
			(1, null);
	`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	defer db.Exec(`drop table test_mysql_time`)

	var n, k time.Time
	var p mysql.Time
	err = db.QueryRow(`select value from test_mysql_time where id = 0`).Scan(&p)
	if err != nil {
		t.Fatalf("scan time error; %v", err)
	}
	err = db.QueryRow(`select value from test_mysql_time where id = 0`).Scan(&n)
	if err != nil {
		t.Fatalf("scan native time error; %v", err)
	}
	if !p.Equal(n) {
		t.Fatalf("scan time not equal when insert; expected %v; got %v", n, p)
	}
	err = db.QueryRow(`select value from test_mysql_time where id = 0`).Scan(mysql.NullTime(&k))
	if err != nil {
		t.Fatalf("scan null time error; %v", err)
	}
	if !k.Equal(n) {
		t.Fatalf("scan time not equal when insert; expected %v; got %v", n, p)
	}

	err = db.QueryRow(`select value from test_mysql_time where id = 1`).Scan(&p)
	if err != nil {
		t.Fatalf("scan time error; %v", err)
	}
	if !p.IsZero() {
		t.Fatalf("invalid time; expected empty got %v", p)
	}

	n = time.Now()
	p.Time = n
	var ok bool
	err = db.QueryRow(`select ? = ?`, p, n).Scan(&ok)
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
	if !ok {
		t.Fatalf("invalid time")
	}

	err = db.QueryRow(`select ? = ?`, mysql.NullTime(&n), n).Scan(&ok)
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
	if !ok {
		t.Fatalf("invalid time")
	}

	p.Time = time.Time{}
	_, err = db.Exec(`insert into test_mysql_time (id, value) values (2, ?)`, p)
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}

	_, err = db.Exec(`insert into test_mysql_time (id, value) values (3, ?)`, mysql.NullTime(new(time.Time)))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
}
