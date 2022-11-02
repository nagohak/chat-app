package database

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/nagohak/chat-app/auth"
)

type Options struct {
	Host, Port, Db, User, Password string
}

func psqlInfo(opt *Options) string {

	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		opt.Host, opt.Port, opt.User, opt.Password, opt.Db)
}

func New(opt *Options, auth auth.Auth) (*sql.DB, error) {
	db, err := sql.Open("postgres", psqlInfo(opt))
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

	var username string
	exists := true
	row := db.QueryRow("SELECT username FROM users WHERE username = ? LIMIT 1", "bob")

	if err := row.Scan(&username); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		exists = false
		err = nil
	}

	if !exists {
		password, _ := auth.GeneratePassword("password")

		sqlStmt = `INSERT into users (id, name, username, password) VALUES
					('` + uuid.New().String() + `', 'Bob', 'bob','` + password + `')`

		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, err
		}
	}

	return db, err
}
