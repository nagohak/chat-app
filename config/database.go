package config

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/nagohak/chat-app/auth"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ".chatdb.db")
	if err != nil {
		return nil, err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS rooms (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		private TINYINT NULL
	)
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(255) NOT NULL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		username VARCHAR(255) NULL,
		password VARCHAR(255) NULL
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	password, _ := auth.GeneratePassword("password")

	sqlStmt = `INSERT into users (id, name, username, password) VALUES
					('` + uuid.New().String() + `', 'Bob', 'bob','` + password + `')`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, err
}
