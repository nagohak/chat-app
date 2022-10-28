package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nagohak/chat-app/auth"
	"github.com/nagohak/chat-app/models"
	"github.com/nagohak/chat-app/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	userRepo    = new(mockUserRepo)
	authService = auth.NewAuth()
	api         = NewApi(userRepo, authService)
)

var user = &repository.User{
	Id:       "1",
	Name:     "tester",
	Username: "tester",
	Password: "123456",
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) AddUser(user models.User) error {
	args := m.Called()
	return args.Error(1)
}
func (m *mockUserRepo) AddDbUser(id uuid.UUID, name, username, password string) (models.DbUser, error) {
	args := m.Called()
	return args.Get(0).(models.DbUser), args.Error(1)
}
func (m *mockUserRepo) RemoveUser(user models.User) error {
	args := m.Called()
	return args.Error(1)
}
func (m *mockUserRepo) FindUserById(id string) (models.User, error) {
	args := m.Called()
	return args.Get(0).(models.User), args.Error(1)
}
func (m *mockUserRepo) GetAllUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) FindUserByUsername(username string) (models.DbUser, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(models.DbUser), args.Error(1)
}

func TestRegistrationOk(t *testing.T) {
	data := []byte(`{
		"name": "` + user.Name + `",
		"username": "` + user.Username + `",
		"password": "` + user.Password + `",
		"confirmation": "` + user.Password + `"
	}`)

	userRepo.On("FindUserByUsername").Once().Return(nil, nil)
	userRepo.On("AddDbUser").Once().Return(user, nil)

	req, _ := http.NewRequest("POST", "/registration", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Registration)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	token := resp.Body.String()
	assert.Greater(t, len(token), 0)
}

func TestRegistrationUserExists(t *testing.T) {
	data := []byte(`{
		"name": "` + user.Name + `",
		"username": "` + user.Username + `",
		"password": "` + user.Password + `",
		"config": "` + user.Password + `"
	}`)

	userRepo.On("FindUserByUsername").Once().Return(user, nil)

	req, _ := http.NewRequest("POST", "/registration", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Registration)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestRegistrationPasswordDontMatch(t *testing.T) {

	data := []byte(`{
		"name": "` + user.Name + `",
		"username": "` + user.Username + `",
		"password": "` + user.Password + `",
		"config": "` + "qwerty" + `"
	}`)

	userRepo.On("FindUserByUsername").Once().Return(user, nil)

	req, _ := http.NewRequest("POST", "/registration", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Registration)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestLogin(t *testing.T) {
	data := []byte(`{
		"username": "` + user.Username + `",
		"password": "` + user.Password + `"
	}`)

	pwd, _ := auth.NewAuth().GeneratePassword(user.Password)
	loginUser := user
	loginUser.Password = pwd
	userRepo.On("FindUserByUsername").Once().Return(loginUser, nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Login)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	token := resp.Body.String()
	assert.Greater(t, len(token), 0)
}

func TestLoginInvalidUsername(t *testing.T) {
	data := []byte(`{
		"username": "` + user.Username + `",
		"password": "` + user.Password + `"
	}`)

	pwd, _ := auth.NewAuth().GeneratePassword(user.Password)
	loginUser := user
	loginUser.Password = pwd
	userRepo.On("FindUserByUsername").Once().Return(nil, nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Login)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestLoginInvalidPassword(t *testing.T) {
	pwd, _ := auth.NewAuth().GeneratePassword(user.Password)
	loginUser := user
	loginUser.Password = pwd

	data := []byte(`{
		"username": "` + user.Username + `",
		"password": "` + user.Password + "1" + `"
	}`)
	userRepo.On("FindUserByUsername").Once().Return(loginUser, nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(data))
	handler := http.HandlerFunc(api.Login)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}
