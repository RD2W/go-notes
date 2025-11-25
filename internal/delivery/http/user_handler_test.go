package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService - мок-объект для UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(username, email, password string) (*model.User, error) {
	args := m.Called(username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id, username, email string) (*model.User, error) {
	args := m.Called(id, username, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers() ([]*model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Подготовка тестовых данных
	tests := []struct {
		name               string
		requestBody        string
		mockCreateUser     func(*MockUserService)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:        "успешное создание пользователя",
			requestBody: createUserRequestJSON("testuser", "test@example.com", "password123"),
			mockCreateUser: func(m *MockUserService) {
				user := createTestUser("1", "testuser", "test@example.com", "password123")
				m.On("CreateUser", "testuser", "test@example.com", "password123").Return(user, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   map[string]interface{}{"id": "1", "username": "testuser", "email": "test@example.com"},
		},
		{
			name:               "ошибка валидации - отсутствует username",
			requestBody:        `{"email": "test@example.com", "password": "password123"}`,
			mockCreateUser:     func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   map[string]interface{}{"error": "Key: 'createUserRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag"},
		},
		{
			name:               "ошибка валидации - отсутствует email",
			requestBody:        `{"username": "testuser", "password": "password123"}`,
			mockCreateUser:     func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   map[string]interface{}{"error": "Key: 'createUserRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag"},
		},
		{
			name:               "ошибка валидации - отсутствует password",
			requestBody:        `{"username": "testuser", "email": "test@example.com"}`,
			mockCreateUser:     func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   map[string]interface{}{"error": "Key: 'createUserRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag"},
		},
		{
			name:        "внутренняя ошибка сервиса",
			requestBody: `{"username": "testuser", "email": "test@example.com", "password": "password123"}`,
			mockCreateUser: func(m *MockUserService) {
				m.On("CreateUser", "testuser", "test@example.com", "password123").Return(nil, assert.AnError)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   map[string]interface{}{"error": "Failed to create user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)
			tt.mockCreateUser(mockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateUser(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Для успешного создания сравниваем только ID, username и email
			if tt.expectedStatusCode == http.StatusCreated {
				assert.Equal(t, "1", response["id"])
				assert.Equal(t, "testuser", response["username"])
				assert.Equal(t, "test@example.com", response["email"])
			} else {
				assert.Equal(t, tt.expectedResponse["error"], response["error"])
			}

			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		userId             string
		mockGetUser        func(*MockUserService)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:   "успешное получение пользователя",
			userId: "1",
			mockGetUser: func(m *MockUserService) {
				user := createTestUser("1", "testuser", "test@example.com", "password123")
				m.On("GetUserByID", "1").Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   map[string]interface{}{"id": "1", "username": "testuser", "email": "test@example.com"},
		},
		{
			name:               "пользователь не найден",
			userId:             "999",
			mockGetUser:        func(m *MockUserService) { m.On("GetUserByID", "999").Return(nil, assert.AnError) },
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   map[string]interface{}{"error": "User not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)
			tt.mockGetUser(mockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodGet, "/users/"+tt.userId, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.AddParam("id", tt.userId)

			handler.GetUser(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatusCode == http.StatusOK {
				assert.Equal(t, tt.expectedResponse["id"], response["id"])
				assert.Equal(t, tt.expectedResponse["username"], response["username"])
				assert.Equal(t, tt.expectedResponse["email"], response["email"])
			} else {
				assert.Equal(t, tt.expectedResponse["error"], response["error"])
			}

			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		userId             string
		requestBody        string
		mockUpdateUser     func(*MockUserService)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:        "успешное обновление пользователя",
			userId:      "1",
			requestBody: updateUserRequestJSON("1", "updateduser", "updated@example.com"),
			mockUpdateUser: func(m *MockUserService) {
				user := createTestUser("1", "updateduser", "updated@example.com", "password123")
				m.On("UpdateUser", "1", "updateduser", "updated@example.com").Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   map[string]interface{}{"id": "1", "username": "updateduser", "email": "updated@example.com"},
		},
		{
			name:               "ошибка валидации - некорректный JSON",
			userId:             "1",
			requestBody:        `{"id": "1", "username":}`,
			mockUpdateUser:     func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   map[string]interface{}{"error": "invalid character '}' looking for beginning of value"},
		},
		{
			name:        "пользователь не найден",
			userId:      "999",
			requestBody: `{"id": "999", "username": "updateduser", "email": "updated@example.com"}`,
			mockUpdateUser: func(m *MockUserService) {
				m.On("UpdateUser", "999", "updateduser", "updated@example.com").Return(nil, assert.AnError)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   map[string]interface{}{"error": "User not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)
			tt.mockUpdateUser(mockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodPut, "/users/"+tt.userId, bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.AddParam("id", tt.userId)

			handler.UpdateUser(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil && tt.expectedResponse["error"] != "invalid character '}' looking for beginning of value" {
				assert.NoError(t, err)
			}

			if tt.expectedStatusCode == http.StatusOK {
				assert.Equal(t, tt.expectedResponse["id"], response["id"])
				assert.Equal(t, tt.expectedResponse["username"], response["username"])
				assert.Equal(t, tt.expectedResponse["email"], response["email"])
			} else {
				if response["error"] != nil {
					assert.Contains(t, response["error"], tt.expectedResponse["error"])
				}
			}

			mockUserService.AssertExpectations(t)
		})
	}
}
func TestUserHandler_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		userId             string
		mockDeleteUser     func(*MockUserService)
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:   "успешное удаление пользователя",
			userId: "1",
			mockDeleteUser: func(m *MockUserService) {
				m.On("DeleteUser", "1").Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "пользователь не найден",
			userId:             "999",
			mockDeleteUser:     func(m *MockUserService) { m.On("DeleteUser", "999").Return(assert.AnError) },
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   map[string]interface{}{"error": "User not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)
			tt.mockDeleteUser(mockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodDelete, "/users/"+tt.userId, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.AddParam("id", tt.userId)

			handler.DeleteUser(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			// Для статуса 204 тело ответа пустое, поэтому не пытаемся его распарсить
			if tt.expectedStatusCode != http.StatusNoContent {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse["error"], response["error"])
			}

			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		mockGetAllUsers    func(*MockUserService)
		expectedStatusCode int
	}{
		{
			name: "успешное получение списка пользователей",
			mockGetAllUsers: func(m *MockUserService) {
				user1 := createTestUser("1", "testuser1", "test1@example.com", "password123")
				user2 := createTestUser("2", "testuser2", "test2@example.com", "password123")
				users := []*model.User{user1, user2}
				m.On("GetAllUsers").Return(users, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "ошибка при получении списка пользователей",
			mockGetAllUsers:    func(m *MockUserService) { m.On("GetAllUsers").Return(nil, assert.AnError) },
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)
			tt.mockGetAllUsers(mockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.GetAllUsers(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedStatusCode == http.StatusOK {
				var response []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
				assert.Equal(t, "1", response[0]["id"])
				assert.Equal(t, "testuser1", response[0]["username"])
				assert.Equal(t, "test1@example.com", response[0]["email"])
				assert.Equal(t, "2", response[1]["id"])
				assert.Equal(t, "testuser2", response[1]["username"])
				assert.Equal(t, "test2@example.com", response[1]["email"])
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Failed to retrieve users", response["error"])
			}

			mockUserService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_CreateUser_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{
			name:               "ошибка валидации - пустой username",
			requestBody:        createUserRequestJSON("", "test@example.com", "password123"),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "ошибка валидации - пустой email",
			requestBody:        createUserRequestJSON("testuser", "", "password123"),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "ошибка валидации - пустой password",
			requestBody:        createUserRequestJSON("testuser", "test@example.com", ""),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "ошибка валидации - все поля пустые",
			requestBody:        createUserRequestJSON("", "", ""),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockUserService)

			handler := NewUserHandler(mockUserService)

			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateUser(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Contains(t, w.Body.String(), "error")
		})
	}
}

// Вспомогательные функции для тестирования

// createUserRequestJSON создает JSON-представление для запроса создания пользователя
func createUserRequestJSON(username, email, password string) string {
	req := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Username: username,
		Email:    email,
		Password: password,
	}

	jsonData, _ := json.Marshal(req)
	return string(jsonData)
}

// updateUserRequestJSON создает JSON-представление для запроса обновления пользователя
func updateUserRequestJSON(id, username, email string) string {
	req := struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}{
		ID:       id,
		Username: username,
		Email:    email,
	}

	jsonData, _ := json.Marshal(req)
	return string(jsonData)
}

// createTestUser создает тестового пользователя с заданными параметрами
func createTestUser(id, username, email, password string) *model.User {
	user, _ := model.NewUser(username, email, password)
	if id != "" {
		user.SetID(id)
	}
	return user
}
