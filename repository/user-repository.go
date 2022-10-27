package repository

import (
	"database/sql"

	"github.com/nagohak/chat-app/models"
)

type user struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (user *user) GetID() string {
	return user.Id
}

func (user *user) GetName() string {
	return user.Name
}

func (user *user) GetUsername() string {
	return user.Username
}

func (user *user) GetPassword() string {
	return user.GetPassword()
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) models.UserRepository {
	return &userRepository{db: db}
}

func (repo *userRepository) AddUser(user models.User) error {
	stmt, err := repo.db.Prepare("INSERT INTO users(id, name) values (?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.GetID(), user.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) RemoveUser(user models.User) error {
	stmt, err := repo.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return err
	}

	stmt.Exec(user.GetID())
	if err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) FindUserById(id string) (models.User, error) {
	row := repo.db.QueryRow("SELECT id, name FROM users WHERE id = ?")

	var user user

	if err := row.Scan(&user.Id, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (repo *userRepository) FindUserByUsername(username string) (models.DbUser, error) {
	row := repo.db.QueryRow("SELECT id, name, username, password FROM users WHERE username = ? LIMIT 1", username)

	var user user

	if err := row.Scan(&user.Id, &user.Name, &user.Username, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (repo *userRepository) GetAllUsers() ([]models.User, error) {
	rows, err := repo.db.Query("SELECT id, name FROM users")
	if err != nil {
		return nil, err
	}

	var users []models.User
	defer rows.Close()

	for rows.Next() {
		var user user
		rows.Scan(&user.Id, &user.Name)
		users = append(users, &user)
	}

	return users, nil
}
