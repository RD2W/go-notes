package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/domain/model"
	userPb "github.com/rd2w/go-notes/pkg/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService - мок-объект для UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(username, email, password string) (*model.User, error) {
	args := m.Called(username, email, password)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id, username, email string) (*model.User, error) {
	args := m.Called(id, username, email)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers() ([]*model.User, error) {
	args := m.Called()
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	return args.Get(0).(*model.User), args.Error(1)
}

func TestNewUserServiceServer(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	assert.NotNil(t, server)
	assert.Equal(t, mockUserService, server.userService)
}

func TestUserServiceServer_CreateUser(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUser := &model.User{}
	expectedUser.SetID("123")
	expectedUser.SetUsername("testuser")
	expectedUser.SetEmail("test@example.com")
	expectedUser.SetCreatedAt(time.Now())
	expectedUser.SetUpdatedAt(time.Now())

	mockUserService.On("CreateUser", req.Username, req.Email, req.Password).Return(expectedUser, nil).Once()

	resp, err := server.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedUser.GetID(), resp.User.Id)
	assert.Equal(t, expectedUser.GetUsername(), resp.User.Username)
	assert.Equal(t, expectedUser.GetEmail(), resp.User.Email)

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_CreateUser_Error(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserService.On("CreateUser", req.Username, req.Email, req.Password).Return((*model.User)(nil), errors.New("user creation failed")).Once()

	resp, err := server.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "user creation failed", err.Error())

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_GetUser(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.GetRequest{
		Id: "123",
	}

	expectedUser := &model.User{}
	expectedUser.SetID("123")
	expectedUser.SetUsername("testuser")
	expectedUser.SetEmail("test@example.com")
	expectedUser.SetCreatedAt(time.Now())
	expectedUser.SetUpdatedAt(time.Now())

	mockUserService.On("GetUserByID", req.Id).Return(expectedUser, nil).Once()

	resp, err := server.GetUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedUser.GetID(), resp.User.Id)
	assert.Equal(t, expectedUser.GetUsername(), resp.User.Username)
	assert.Equal(t, expectedUser.GetEmail(), resp.User.Email)

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_GetUser_Error(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.GetRequest{
		Id: "123",
	}

	mockUserService.On("GetUserByID", req.Id).Return((*model.User)(nil), errors.New("user not found")).Once()

	resp, err := server.GetUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "user not found", err.Error())

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_UpdateUser(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.UpdateUserRequest{
		Id:       "123",
		Username: "updateduser",
		Email:    "updated@example.com",
	}

	expectedUser := &model.User{}
	expectedUser.SetID("123")
	expectedUser.SetUsername("updateduser")
	expectedUser.SetEmail("updated@example.com")
	expectedUser.SetCreatedAt(time.Now())
	expectedUser.SetUpdatedAt(time.Now())

	mockUserService.On("UpdateUser", req.Id, req.Username, req.Email).Return(expectedUser, nil).Once()

	resp, err := server.UpdateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedUser.GetID(), resp.User.Id)
	assert.Equal(t, expectedUser.GetUsername(), resp.User.Username)
	assert.Equal(t, expectedUser.GetEmail(), resp.User.Email)

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_UpdateUser_Error(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.UpdateUserRequest{
		Id:       "123",
		Username: "updateduser",
		Email:    "updated@example.com",
	}

	mockUserService.On("UpdateUser", req.Id, req.Username, req.Email).Return((*model.User)(nil), errors.New("user update failed")).Once()

	resp, err := server.UpdateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "user update failed", err.Error())

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_DeleteUser(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.GetRequest{
		Id: "123",
	}

	mockUserService.On("DeleteUser", req.Id).Return(nil).Once()

	resp, err := server.DeleteUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user deleted successfully", resp.Message)

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_DeleteUser_Error(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.GetRequest{
		Id: "123",
	}

	mockUserService.On("DeleteUser", req.Id).Return(errors.New("user deletion failed")).Once()

	resp, err := server.DeleteUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "user deletion failed", err.Error())

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_ListUsers(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.Empty{}

	expectedUser1 := &model.User{}
	expectedUser1.SetID("123")
	expectedUser1.SetUsername("testuser1")
	expectedUser1.SetEmail("test1@example.com")
	expectedUser1.SetCreatedAt(time.Now())
	expectedUser1.SetUpdatedAt(time.Now())

	expectedUser2 := &model.User{}
	expectedUser2.SetID("456")
	expectedUser2.SetUsername("testuser2")
	expectedUser2.SetEmail("test2@example.com")
	expectedUser2.SetCreatedAt(time.Now())
	expectedUser2.SetUpdatedAt(time.Now())

	expectedUsers := []*model.User{expectedUser1, expectedUser2}

	mockUserService.On("GetAllUsers").Return(expectedUsers, nil).Once()

	resp, err := server.ListUsers(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, expectedUsers[0].GetID(), resp.Users[0].Id)
	assert.Equal(t, expectedUsers[0].GetUsername(), resp.Users[0].Username)
	assert.Equal(t, expectedUsers[0].GetEmail(), resp.Users[0].Email)
	assert.Equal(t, expectedUsers[1].GetID(), resp.Users[1].Id)
	assert.Equal(t, expectedUsers[1].GetUsername(), resp.Users[1].Username)
	assert.Equal(t, expectedUsers[1].GetEmail(), resp.Users[1].Email)

	mockUserService.AssertExpectations(t)
}

func TestUserServiceServer_ListUsers_Error(t *testing.T) {
	mockUserService := new(MockUserService)
	server := NewUserServiceServer(mockUserService)

	ctx := context.Background()
	req := &userPb.Empty{}

	mockUserService.On("GetAllUsers").Return([]*model.User{}, errors.New("failed to get users")).Once()

	resp, err := server.ListUsers(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "failed to get users", err.Error())

	mockUserService.AssertExpectations(t)
}
