package mysql_test

import (
	"log"
	"testing"

	"github.com/acoshift/mysql"
)

func TestJSON(t *testing.T) {
	db := open(t)

	_, err := db.Exec(`drop table if exists test_mysql_json;`)
	if err != nil {
		log.Fatalf("prepare table error; %v", err)
	}
	_, err = db.Exec(`
	create table test_mysql_json (
		id int primary key,
		value longblob
	);`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	defer db.Exec(`drop table test_mysql_json`)

	var obj struct {
		A string
		B int
	}

	obj.A = "test"
	obj.B = 7

	_, err = db.Exec(`
		insert into test_mysql_json (id, value)
		values (1, ?)
	`, mysql.JSON(&obj))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`
		select value
		from test_mysql_json
		where id = 1
	`).Scan(mysql.JSON(&obj))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
	if obj.A != "test" || obj.B != 7 {
		t.Fatal("invalid object scanner")
	}

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`select null`).Scan(mysql.JSON(&obj))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
}
