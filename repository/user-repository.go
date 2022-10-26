package repository

import (
	"database/sql"

	"github.com/nagohak/chat-app/models"
)

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (user *User) GetID() string {
	return user.Id
}

func (user *User) GetName() string {
	return user.Name
}

type UserRepository struct {
	Db *sql.DB
}

func (repo *UserRepository) AddUser(user models.User) error {
	stmt, err := repo.Db.Prepare("INSERT INTO users(id, name) values (?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.GetID(), user.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (repo *UserRepository) RemoveUser(user models.User) error {
	stmt, err := repo.Db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return err
	}

	stmt.Exec(user.GetID())
	if err != nil {
		return err
	}

	return nil
}

func (repo *UserRepository) FindUserById(id string) (models.User, error) {
	row := repo.Db.QueryRow("SELECT id, name FROM users WHERE id = ?")

	var user User

	if err := row.Scan(&user.Id, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) FindUserByUsername(username string) (*User, error) {
	row := repo.Db.QueryRow("SELECT id, name, username, password FROM users WHERE username = ? LIMIT 1", username)

	var user User

	if err := row.Scan(&user.Id, &user.Name, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := repo.Db.Query("SELECT id, name FROM users")
	if err != nil {
		return nil, err
	}

	var users []models.User
	defer rows.Close()

	for rows.Next() {
		var user User
		rows.Scan(&user.Id, &user.Name)
		users = append(users, &user)
	}

	return users, nil
}
