package models

type User interface {
	GetID() string
	GetName() string
}

type UserRepository interface {
	AddUser(user User) error
	RemoveUser(user User) error
	FindUserById(id string) (User, error)
	GetAllUsers() ([]User, error)
}
