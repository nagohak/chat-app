package repository

import (
	"database/sql"

	"github.com/google/uuid"
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

func (user *User) GetUsername() string {
	return user.Username
}

func (user *User) GetPassword() string {
	return user.Password
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) models.UserRepository {
	return &userRepository{db: db}
}

func (repo *userRepository) AddUser(user models.User) error {
	stmt, err := repo.db.Prepare("INSERT INTO users(id, name) values ($1, $2)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.GetID(), user.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) AddDbUser(id uuid.UUID, name, username, password string) (models.DbUser, error) {
	user := &User{
		Id:       id.String(),
		Name:     name,
		Username: username,
		Password: password,
	}

	stmt, err := repo.db.Prepare("INSERT INTO users(id, name, username, password) values ($1, $2, $3, $4)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(user.Id, user.Name, user.Username, user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *userRepository) RemoveUser(user models.User) error {
	stmt, err := repo.db.Prepare("DELETE FROM users WHERE id = $1")
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
	row := repo.db.QueryRow("SELECT id, name FROM users WHERE id = $1")

	var user User

	if err := row.Scan(&user.Id, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (repo *userRepository) FindUserByUsername(username string) (models.DbUser, error) {
	row := repo.db.QueryRow("SELECT id, name, username, password FROM users WHERE username = $1 LIMIT 1", username)

	var user User

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
		var user User
		rows.Scan(&user.Id, &user.Name)
		users = append(users, &user)
	}

	return users, nil
}
