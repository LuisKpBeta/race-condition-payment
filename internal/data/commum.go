package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectToDabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./mydb.db")
	if err != nil {
		panic("error on connection with database")
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS clients (id INTEGER PRIMARY KEY, name VARCHAR(256) NULL);")
	if err != nil {
		panic("error on creating users table")
	}
	return db
}
func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
