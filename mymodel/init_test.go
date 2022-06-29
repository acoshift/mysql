package mymodel_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func open(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "root@tcp(127.0.0.1:3306)/mysql"
	}
	db, err := sql.Open("mysql", dbURL+"?parseTime=true")
	if err != nil {
		t.Fatalf("open database connection error; %v", err)
	}
	return db
}
