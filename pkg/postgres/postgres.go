package postgres

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

func MigrationUp(db *sql.DB) error {
	m, err := newMigrate(db)
	if err != nil {
		return nil
	}

	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}

	return nil
}

func MigrationDown(db *sql.DB) error {
	m, err := newMigrate(db)
	if err != nil {
		return nil
	}

	if err = m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}

	return nil
}

func newMigrate(db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres",
		driver,
	)
}
