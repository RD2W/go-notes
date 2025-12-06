package grpc

import (
	"context"
	"log"

	"github.com/rd2w/go-notes/internal/domain/service"
	userPb "github.com/rd2w/go-notes/pkg/proto/user"
)

// UserServiceServer реализует gRPC-сервер для сервиса пользователей
type UserServiceServer struct {
	userPb.UnimplementedUserServiceServer
	userService service.UserService
}

// NewUserServiceServer создает новый экземпляр gRPC-сервера для пользователей
func NewUserServiceServer(userService service.UserService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
	}
}

// CreateUser создает нового пользователя
func (s *UserServiceServer) CreateUser(ctx context.Context, req *userPb.CreateUserRequest) (*userPb.UserResponse, error) {
	user, err := s.userService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &userPb.UserResponse{
		User: &userPb.User{
			Id:        user.GetID(),
			Username:  user.GetUsername(),
			Email:     user.GetEmail(),
			CreatedAt: user.GetCreatedAt().Unix(),
			UpdatedAt: user.GetUpdatedAt().Unix(),
		},
	}, nil
}

// GetUser возвращает пользователя по ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *userPb.GetRequest) (*userPb.UserResponse, error) {
	user, err := s.userService.GetUserByID(req.Id)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return nil, err
	}

	return &userPb.UserResponse{
		User: &userPb.User{
			Id:        user.GetID(),
			Username:  user.GetUsername(),
			Email:     user.GetEmail(),
			CreatedAt: user.GetCreatedAt().Unix(),
			UpdatedAt: user.GetUpdatedAt().Unix(),
		},
	}, nil
}

// UpdateUser обновляет пользователя
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *userPb.UpdateUserRequest) (*userPb.UserResponse, error) {
	user, err := s.userService.UpdateUser(req.Id, req.Username, req.Email)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, err
	}

	return &userPb.UserResponse{
		User: &userPb.User{
			Id:        user.GetID(),
			Username:  user.GetUsername(),
			Email:     user.GetEmail(),
			CreatedAt: user.GetCreatedAt().Unix(),
			UpdatedAt: user.GetUpdatedAt().Unix(),
		},
	}, nil
}

// DeleteUser удаляет пользователя
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *userPb.GetRequest) (*userPb.SuccessResponse, error) {
	err := s.userService.DeleteUser(req.Id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return nil, err
	}

	return &userPb.SuccessResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

// ListUsers возвращает список всех пользователей
func (s *UserServiceServer) ListUsers(ctx context.Context, req *userPb.Empty) (*userPb.UsersListResponse, error) {
	users, err := s.userService.GetAllUsers()
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return nil, err
	}

	protoUsers := make([]*userPb.User, len(users))
	for i, user := range users {
		protoUsers[i] = &userPb.User{
			Id:        user.GetID(),
			Username:  user.GetUsername(),
			Email:     user.GetEmail(),
			CreatedAt: user.GetCreatedAt().Unix(),
			UpdatedAt: user.GetUpdatedAt().Unix(),
		}
	}

	return &userPb.UsersListResponse{
		Users: protoUsers,
	}, nil
}
