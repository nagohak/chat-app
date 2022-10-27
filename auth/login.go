package auth

type contextKey string

const UserContextKey = contextKey("user")

type NewUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (user *NewUser) GetID() string {
	return user.Id
}

func (user *NewUser) GetName() string {
	return user.Name
}
