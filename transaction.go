package sorm

import (
	"database/sql"
)

//具体的事务操作在f()中执行
func Transaction(db *sql.DB, f func(Dba) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		p := recover()
		if err != nil {
			if rerr := tx.Rollback(); err != nil {
				panic(rerr)
			}
			return
		}
		if p != nil {
			if rerr := tx.Rollback(); err != nil {
				panic(rerr)
			}
			return
		}
		if cerr := tx.Commit(); err != nil {
			panic(cerr)
		}
	}()
	err = f(tx)
	return err
}
