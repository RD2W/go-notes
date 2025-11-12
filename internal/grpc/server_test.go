package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	"github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/rd2w/go-notes/pkg/proto/user"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockRepository - тестовая реализация репозитория
type MockRepository struct {
	entities map[string]repository.Entity
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		entities: make(map[string]repository.Entity),
	}
}

func (m *MockRepository) Save(entity repository.Entity) {
	m.entities[entity.GetID()] = entity
}

func (m *MockRepository) GetAllNotes() []*model.Note {
	var notes []*model.Note
	for _, entity := range m.entities {
		if noteEntity, ok := entity.(*model.Note); ok {
			notes = append(notes, noteEntity)
		}
	}
	return notes
}

func (m *MockRepository) GetNotesCount() int {
	count := 0
	for _, entity := range m.entities {
		if _, ok := entity.(*model.Note); ok {
			count++
		}
	}
	return count
}

func (m *MockRepository) GetNewNotes(lastIndex int) []*model.Note {
	return m.GetAllNotes()
}

func (m *MockRepository) GetAllByType(entityType string) []repository.Entity {
	var entities []repository.Entity
	for _, entity := range m.entities {
		if entity.GetType() == entityType {
			entities = append(entities, entity)
		}
	}
	return entities
}

func (m *MockRepository) GetByID(entityType, id string) repository.Entity {
	entity := m.entities[id]
	if entity != nil && entity.GetType() == entityType {
		return entity
	}
	return nil
}

func (m *MockRepository) DeleteByID(entityType, id string) bool {
	entity := m.entities[id]
	if entity != nil && entity.GetType() == entityType {
		delete(m.entities, id)
		return true
	}
	return false
}

// MockTokenManager - тестовая реализация TokenManager
type MockTokenManager struct {
	shouldFailGenerateTokens bool
	shouldFailValidateToken  bool
	shouldFailRefresh        bool
	shouldFailLogout         bool
}

func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{}
}

func (m *MockTokenManager) GenerateTokens(username string) (string, string, error) {
	if m.shouldFailGenerateTokens {
		return "", "", fmt.Errorf("ошибка генерации токенов")
	}
	return "access_token", "refresh_token", nil
}

func (m *MockTokenManager) RefreshTokens(refreshToken string) (string, string, error) {
	if m.shouldFailRefresh {
		return "", "", fmt.Errorf("ошибка обновления токенов")
	}
	return "new_access_token", "new_refresh_token", nil
}

func (m *MockTokenManager) ValidateAccessToken(tokenString string) (*auth.TokenClaims, error) {
	if m.shouldFailValidateToken {
		return nil, fmt.Errorf("токен недействителен")
	}
	return &auth.TokenClaims{
		Username: "testuser",
		TokenID:  "test_token_id",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}, nil
}

func (m *MockTokenManager) Logout(refreshToken string) error {
	if m.shouldFailLogout {
		return fmt.Errorf("ошибка при выходе")
	}
	return nil
}

func (m *MockTokenManager) GetJWTExpiration() time.Duration {
	return 15 * time.Minute
}

func (m *MockTokenManager) GetJWTExpirationSeconds() int64 {
	return 900 // 15 минут в секундах
}

func TestNewServer(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	assert.NotNil(t, server)
	assert.Equal(t, mockRepo, server.repo)
	// tokenManager не может быть напрямую проверен, так как это интерфейс
}

func TestCreateNote(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &note.CreateNoteRequest{
		Title:   "Test Note",
		Content: "Test Content",
	}

	resp, err := server.CreateNote(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Note)
	assert.Equal(t, "Test Note", resp.Note.Title)
	assert.Equal(t, "Test Content", resp.Note.Content)
	assert.NotEmpty(t, resp.Note.Id)

	// Проверяем, что заметка была сохранена в репозитории
	entity := mockRepo.GetByID("note", resp.Note.Id)
	assert.NotNil(t, entity)
	assert.IsType(t, &model.Note{}, entity)
}

func TestGetNote(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестовую заметку
	testNote := model.NewNote("Test Title", "Test Content")
	mockRepo.Save(testNote)

	req := &note.GetRequest{
		Id: testNote.GetID(),
	}

	resp, err := server.GetNote(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Note)
	assert.Equal(t, testNote.GetID(), resp.Note.Id)
	assert.Equal(t, testNote.GetTitle(), resp.Note.Title)
	assert.Equal(t, testNote.GetContent(), resp.Note.Content)
}

func TestGetNoteNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &note.GetRequest{
		Id: "nonexistent-id",
	}

	resp, err := server.GetNote(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateNote(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестовую заметку
	testNote := model.NewNote("Old Title", "Old Content")
	mockRepo.Save(testNote)

	req := &note.UpdateNoteRequest{
		Id:      testNote.GetID(),
		Title:   "New Title",
		Content: "New Content",
	}

	resp, err := server.UpdateNote(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Note)
	assert.Equal(t, testNote.GetID(), resp.Note.Id)
	assert.Equal(t, "New Title", resp.Note.Title)
	assert.Equal(t, "New Content", resp.Note.Content)

	// Проверяем, что заметка была обновлена в репозитории
	entity := mockRepo.GetByID("note", testNote.GetID())
	assert.NotNil(t, entity)
	assert.IsType(t, &model.Note{}, entity)
	assert.Equal(t, "New Title", entity.(*model.Note).GetTitle())
	assert.Equal(t, "New Content", entity.(*model.Note).GetContent())
}

func TestUpdateNoteNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &note.UpdateNoteRequest{
		Id:      "nonexistent-id",
		Title:   "New Title",
		Content: "New Content",
	}

	resp, err := server.UpdateNote(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteNote(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестовую заметку
	testNote := model.NewNote("Test Title", "Test Content")
	mockRepo.Save(testNote)

	req := &note.GetRequest{
		Id: testNote.GetID(),
	}

	resp, err := server.DeleteNote(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "note deleted successfully", resp.Message)

	// Проверяем, что заметка была удалена из репозитория
	entity := mockRepo.GetByID("note", testNote.GetID())
	assert.Nil(t, entity)
}

func TestDeleteNoteNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &note.GetRequest{
		Id: "nonexistent-id",
	}

	resp, err := server.DeleteNote(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListNotes(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем несколько тестовых заметок
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")
	mockRepo.Save(note1)
	mockRepo.Save(note2)

	req := &note.Empty{}

	resp, err := server.ListNotes(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Notes, 2)

	// Проверяем, что заметки содержатся в ответе
	var foundNote1, foundNote2 bool
	for _, n := range resp.Notes {
		if n.Id == note1.GetID() {
			assert.Equal(t, "Note 1", n.Title)
			foundNote1 = true
		}
		if n.Id == note2.GetID() {
			assert.Equal(t, "Note 2", n.Title)
			foundNote2 = true
		}
	}
	assert.True(t, foundNote1)
	assert.True(t, foundNote2)
}

func TestCreateUser(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &user.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := server.CreateUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.User)
	assert.Equal(t, "testuser", resp.User.Username)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.NotEmpty(t, resp.User.Id)

	// Проверяем, что пользователь был сохранен в репозитории
	entity := mockRepo.GetByID("user", resp.User.Id)
	assert.NotNil(t, entity)
	assert.IsType(t, &model.User{}, entity)
}

func TestGetUser(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестового пользователя
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(testUser)

	req := &user.GetRequest{
		Id: testUser.GetID(),
	}

	resp, err := server.GetUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.User)
	assert.Equal(t, testUser.GetID(), resp.User.Id)
	assert.Equal(t, testUser.GetUsername(), resp.User.Username)
	assert.Equal(t, testUser.GetEmail(), resp.User.Email)
}

func TestGetUserNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &user.GetRequest{
		Id: "nonexistent-id",
	}

	resp, err := server.GetUser(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateUser(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестового пользователя
	testUser, err := model.NewUser("olduser", "old@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(testUser)

	req := &user.UpdateUserRequest{
		Id:       testUser.GetID(),
		Username: "newuser",
		Email:    "new@example.com",
	}

	resp, err := server.UpdateUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.User)
	assert.Equal(t, testUser.GetID(), resp.User.Id)
	assert.Equal(t, "newuser", resp.User.Username)
	assert.Equal(t, "new@example.com", resp.User.Email)

	// Проверяем, что пользователь был обновлен в репозитории
	entity := mockRepo.GetByID("user", testUser.GetID())
	assert.NotNil(t, entity)
	assert.IsType(t, &model.User{}, entity)
	assert.Equal(t, "newuser", entity.(*model.User).GetUsername())
	assert.Equal(t, "new@example.com", entity.(*model.User).GetEmail())
}

func TestUpdateUserNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &user.UpdateUserRequest{
		Id:       "nonexistent-id",
		Username: "newuser",
		Email:    "new@example.com",
	}

	resp, err := server.UpdateUser(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteUser(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестового пользователя
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(testUser)

	req := &user.GetRequest{
		Id: testUser.GetID(),
	}

	resp, err := server.DeleteUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user deleted successfully", resp.Message)

	// Проверяем, что пользователь был удален из репозитория
	entity := mockRepo.GetByID("user", testUser.GetID())
	assert.Nil(t, entity)
}

func TestDeleteUserNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &user.GetRequest{
		Id: "nonexistent-id",
	}

	resp, err := server.DeleteUser(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListUsers(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем несколько тестовых пользователей
	user1, err := model.NewUser("user1", "user1@example.com", "password1")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	user2, err := model.NewUser("user2", "user2@example.com", "password2")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(user1)
	mockRepo.Save(user2)

	req := &user.Empty{}

	resp, err := server.ListUsers(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 2)

	// Проверяем, что пользователи содержатся в ответе
	var foundUser1, foundUser2 bool
	for _, u := range resp.Users {
		if u.Id == user1.GetID() {
			assert.Equal(t, "user1", u.Username)
			foundUser1 = true
		}
		if u.Id == user2.GetID() {
			assert.Equal(t, "user2", u.Username)
			foundUser2 = true
		}
	}
	assert.True(t, foundUser1)
	assert.True(t, foundUser2)
}

func TestLogin(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестового пользователя
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(testUser)

	req := &authpb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	resp, err := server.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestLoginNotFound(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	resp, err := server.Login(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestLoginInvalidPassword(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	// Создаем тестового пользователя
	testUser, err := model.NewUser("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	mockRepo.Save(testUser)

	req := &authpb.LoginRequest{
		Username: "testuser",
		Password: "invalid_password",
	}

	resp, err := server.Login(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestLogout(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}

	resp, err := server.Logout(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "успешный выход", resp.Message)
}

func TestLogoutError(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	mockTokenManager.shouldFailLogout = true
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.LogoutRequest{
		RefreshToken: "invalid_refresh_token",
	}

	resp, err := server.Logout(context.Background(), req)

	assert.NoError(t, err) // Logout не возвращает ошибку, даже если токен невалиден
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Message)
}

func TestRefresh(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.RefreshRequest{
		RefreshToken: "valid_refresh_token",
	}

	resp, err := server.Refresh(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestRefreshError(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	mockTokenManager.shouldFailRefresh = true
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.RefreshRequest{
		RefreshToken: "invalid_refresh_token",
	}

	resp, err := server.Refresh(context.Background(), req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestValidateToken(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.ValidateTokenRequest{
		Token: "valid_access_token",
	}

	resp, err := server.ValidateToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Valid)
	assert.Equal(t, "testuser", resp.Username)
	assert.Empty(t, resp.ErrorMessage) // При успешной валидации errorMessage должно быть пустым
}

func TestValidateTokenError(t *testing.T) {
	mockRepo := NewMockRepository()
	mockTokenManager := NewMockTokenManager()
	mockTokenManager.shouldFailValidateToken = true
	server := NewServer(mockRepo, mockTokenManager)

	req := &authpb.ValidateTokenRequest{
		Token: "invalid_access_token",
	}

	resp, err := server.ValidateToken(context.Background(), req)

	assert.NoError(t, err) // ValidateToken не возвращает ошибку gRPC, а возвращает информацию в ответе
	assert.NotNil(t, resp)
	assert.False(t, resp.Valid)
	assert.NotEmpty(t, resp.ErrorMessage)
}
