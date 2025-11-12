package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateUser(t *testing.T) {
	repo := NewMockRepository()
	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/users", handler.CreateUser)

	// Подготовка тестовых данных
	requestBody := createUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseUser model.User
	err := json.Unmarshal(w.Body.Bytes(), &responseUser)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", responseUser.GetUsername())
	assert.Equal(t, "test@example.com", responseUser.GetEmail())
	assert.NotEmpty(t, responseUser.GetID())
}

func TestGetUser(t *testing.T) {
	// Создаем мок-репозиторий с тестовым пользователем
	repo := NewMockRepository()
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)
	repo.Save(testUser)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/users/:id", handler.GetUser)

	req, _ := http.NewRequest(http.MethodGet, "/api/users/"+testUser.GetID(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseUser model.User
	err = json.Unmarshal(w.Body.Bytes(), &responseUser)
	assert.NoError(t, err)
	assert.Equal(t, testUser.GetID(), responseUser.GetID())
	assert.Equal(t, testUser.GetUsername(), responseUser.GetUsername())
	assert.Equal(t, testUser.GetEmail(), responseUser.GetEmail())
}

func TestGetUserNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/users/:id", handler.GetUser)

	req, _ := http.NewRequest(http.MethodGet, "/api/users/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUser(t *testing.T) {
	// Создаем мок-репозиторий с тестовым пользователем
	repo := NewMockRepository()
	testUser, err := model.NewUser("originaluser", "original@example.com", "password123")
	assert.NoError(t, err)
	repo.Save(testUser)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/users/:id", handler.UpdateUser)

	// Подготовка обновленных данных
	updatedData := map[string]interface{}{
		"username": "updateduser",
		"email":    "updated@example.com",
	}
	jsonData, _ := json.Marshal(updatedData)

	req, _ := http.NewRequest(http.MethodPut, "/api/users/"+testUser.GetID(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseUser model.User
	err = json.Unmarshal(w.Body.Bytes(), &responseUser)
	assert.NoError(t, err)
	assert.Equal(t, testUser.GetID(), responseUser.GetID())
	assert.Equal(t, "updateduser", responseUser.GetUsername())
	assert.Equal(t, "updated@example.com", responseUser.GetEmail())
}

func TestUpdateUserNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/users/:id", handler.UpdateUser)

	updatedData := map[string]interface{}{
		"username": "updateduser",
		"email":    "updated@example.com",
	}
	jsonData, _ := json.Marshal(updatedData)

	req, _ := http.NewRequest(http.MethodPut, "/api/users/nonexistent", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteUser(t *testing.T) {
	// Создаем мок-репозиторий с тестовым пользователем
	repo := NewMockRepository()
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)
	repo.Save(testUser)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/users/:id", handler.DeleteUser)

	req, _ := http.NewRequest(http.MethodDelete, "/api/users/"+testUser.GetID(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверяем, что пользователь действительно удален
	assert.Nil(t, repo.GetByID("user", testUser.GetID()))
}

func TestDeleteUserNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/users/:id", handler.DeleteUser)

	req, _ := http.NewRequest(http.MethodDelete, "/api/users/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAllUsers(t *testing.T) {
	// Создаем мок-репозиторий с несколькими пользователями
	repo := NewMockRepository()
	user1, err := model.NewUser("user1", "user1@example.com", "password123")
	assert.NoError(t, err)
	user2, err := model.NewUser("user2", "user2@example.com", "password456")
	assert.NoError(t, err)
	repo.Save(user1)
	repo.Save(user2)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/users", handler.GetAllUsers)

	req, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseUsers []*model.User
	err = json.Unmarshal(w.Body.Bytes(), &responseUsers)
	assert.NoError(t, err)
	assert.Len(t, responseUsers, 2)

	// Проверяем, что оба пользователя присутствуют в ответе
	foundUser1 := false
	foundUser2 := false
	for _, user := range responseUsers {
		if user.GetID() == user1.GetID() {
			foundUser1 = true
			assert.Equal(t, "user1", user.GetUsername())
			assert.Equal(t, "user1@example.com", user.GetEmail())
		}
		if user.GetID() == user2.GetID() {
			foundUser2 = true
			assert.Equal(t, "user2", user.GetUsername())
			assert.Equal(t, "user2@example.com", user.GetEmail())
		}
	}
	assert.True(t, foundUser1)
	assert.True(t, foundUser2)
}

func TestLogin(t *testing.T) {
	// Создаем мок-репозиторий с тестовым пользователем
	repo := NewMockRepository()
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)
	repo.Save(testUser)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/login", handler.Login)

	// Подготовка данных для входа
	loginData := loginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response loginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
}

func TestLoginInvalidCredentials(t *testing.T) {
	// Создаем мок-репозиторий с тестовым пользователем
	repo := NewMockRepository()
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)
	repo.Save(testUser)

	handler := NewUserHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/login", handler.Login)

	// Подготовка неверных данных для входа
	loginData := loginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
