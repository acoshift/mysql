package mysql_test

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/acoshift/mysql"
)

func TestTx(t *testing.T) {
	db := open(t)
	_, err := db.Exec(`drop table if exists test_mysql_tx;`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	_, err = db.Exec(`
		create table test_mysql_tx (
			id int primary key,
			value int
		);
	`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	_, err = db.Exec(`
		insert into test_mysql_tx (
			id, value
		) values
			(0, 0),
			(1, 0),
			(2, 0);
	`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	defer db.Exec(`drop table test_mysql_tx`)
	db.SetMaxOpenConns(30)

	opts := &mysql.TxOptions{}

	deposit := func(balance int) error {
		return mysql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("deposit", balance)
			var acc0, acc1 int
			err = tx.QueryRow(`select value from test_mysql_tx where id = 0`).Scan(&acc0)
			if err != nil {
				return err
			}
			err = tx.QueryRow(`select value from test_mysql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 0`, acc0-balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 1`, acc1+balance)
			if err != nil {
				return err
			}
			return nil
		})
	}
	withdraw := func(balance int) error {
		return mysql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("withdraw", balance)
			var acc0, acc1 int
			err = tx.QueryRow(`select value from test_mysql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			if acc1 < balance {
				return fmt.Errorf("not enough balance to withdraw")
			}
			err = tx.QueryRow(`select value from test_mysql_tx where id = 0`).Scan(&acc0)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 0`, acc0+balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 1`, acc1-balance)
			if err != nil {
				return err
			}
			return nil
		})
	}
	transfer := func(balance int) error {
		return mysql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("transfer", balance)
			var acc1, acc2 int
			err = tx.QueryRow(`select value from test_mysql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			if acc1 < balance {
				return fmt.Errorf("not enough balance to transfer")
			}
			err = tx.QueryRow(`select value from test_mysql_tx where id = 2`).Scan(&acc2)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 1`, acc1-balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_mysql_tx set value = ? where id = 2`, acc2+balance)
			if err != nil {
				return err
			}
			return nil
		})
	}

	wg := sync.WaitGroup{}
	rand.Seed(time.Now().Unix())
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			var err error
			k := rand.Intn(3)
			if k == 0 {
				err = deposit(rand.Intn(100000))
			} else if k == 1 {
				err = withdraw(rand.Intn(100000))
			} else {
				err = transfer(rand.Intn(100000))
			}
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	var result int
	err = db.QueryRow(`select sum(value) from test_mysql_tx`).Scan(&result)
	if err != nil {
		t.Fatalf("query result error; %v", err)
	}
	if result != 0 {
		t.Fatalf("expected sum all value to be 0; got %d", result)
	}
}
