package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var (
	connectionString string
	db               *sql.DB
)

func Connect(connString string) error {
	var err error
	db, err = sql.Open("sqlite3", connString)
	err = db.Ping()
	if err != nil {
		return errors.Wrapf(err, "while opening db (%s)", connString)
	}

	connectionString = connString
	return nil
}

func Close() {
	db.Close()
	connectionString = ""
}

// ConnectionString gets the connection string currently in use.
func ConnectionString() string {
	return connectionString
}
