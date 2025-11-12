package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC-сервер для сервиса заметок
type Server struct {
	note.UnimplementedNotesServiceServer
	user.UnimplementedUserServiceServer
	repo repository.Repository
}

// NewServer создает новый экземпляр gRPC-сервера
func NewServer(r repository.Repository) *Server {
	return &Server{
		repo: r,
	}
}

// CreateNote создает новую заметку
func (s *Server) CreateNote(ctx context.Context, req *note.CreateNoteRequest) (*note.NoteResponse, error) {
	newNote := model.NewNote(req.Title, req.Content)
	s.repo.Save(newNote)

	return &note.NoteResponse{
		Note: &note.Note{
			Id:        newNote.GetID(),
			Title:     newNote.GetTitle(),
			Content:   newNote.GetContent(),
			CreatedAt: newNote.GetCreatedAt().Unix(),
			UpdatedAt: newNote.GetUpdatedAt().Unix(),
		},
	}, nil
}

// GetNote возвращает заметку по ID
func (s *Server) GetNote(ctx context.Context, req *note.GetRequest) (*note.NoteResponse, error) {
	entity := s.repo.GetByID("note", req.Id)
	if entity == nil {
		return nil, status.Error(codes.NotFound, "note not found")
	}

	noteEntity, ok := entity.(*model.Note)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast entity to note")
	}

	return &note.NoteResponse{
		Note: &note.Note{
			Id:        noteEntity.GetID(),
			Title:     noteEntity.GetTitle(),
			Content:   noteEntity.GetContent(),
			CreatedAt: noteEntity.GetCreatedAt().Unix(),
			UpdatedAt: noteEntity.GetUpdatedAt().Unix(),
		},
	}, nil
}

// UpdateNote обновляет заметку
func (s *Server) UpdateNote(ctx context.Context, req *note.UpdateNoteRequest) (*note.NoteResponse, error) {
	entity := s.repo.GetByID("note", req.Id)
	if entity == nil {
		return nil, status.Error(codes.NotFound, "note not found")
	}

	noteEntity, ok := entity.(*model.Note)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast entity to note")
	}

	noteEntity.SetTitle(req.Title)
	noteEntity.SetContent(req.Content)

	s.repo.Save(noteEntity)

	return &note.NoteResponse{
		Note: &note.Note{
			Id:        noteEntity.GetID(),
			Title:     noteEntity.GetTitle(),
			Content:   noteEntity.GetContent(),
			CreatedAt: noteEntity.GetCreatedAt().Unix(),
			UpdatedAt: noteEntity.GetUpdatedAt().Unix(),
		},
	}, nil
}

// DeleteNote удаляет заметку
func (s *Server) DeleteNote(ctx context.Context, req *note.GetRequest) (*note.SuccessResponse, error) {
	deleted := s.repo.DeleteByID("note", req.Id)
	if !deleted {
		return nil, status.Error(codes.NotFound, "note not found")
	}

	return &note.SuccessResponse{
		Success: true,
		Message: "note deleted successfully",
	}, nil
}

// ListNotes возвращает список всех заметок
func (s *Server) ListNotes(ctx context.Context, req *note.Empty) (*note.NotesListResponse, error) {
	notes := s.repo.GetAllNotes()
	protoNotes := make([]*note.Note, len(notes))

	for i, noteEntity := range notes {
		protoNotes[i] = &note.Note{
			Id:        noteEntity.GetID(),
			Title:     noteEntity.GetTitle(),
			Content:   noteEntity.GetContent(),
			CreatedAt: noteEntity.GetCreatedAt().Unix(),
			UpdatedAt: noteEntity.GetUpdatedAt().Unix(),
		}
	}

	return &note.NotesListResponse{
		Notes: protoNotes,
	}, nil
}

// CreateUser создает нового пользователя
func (s *Server) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	newUser, err := model.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create user: %v", err))
	}

	s.repo.Save(newUser)

	return &user.UserResponse{
		User: &user.User{
			Id:        newUser.GetID(),
			Username:  newUser.GetUsername(),
			Email:     newUser.GetEmail(),
			CreatedAt: newUser.GetCreatedAt().Unix(),
			UpdatedAt: newUser.GetUpdatedAt().Unix(),
		},
	}, nil
}

// GetUser возвращает пользователя по ID
func (s *Server) GetUser(ctx context.Context, req *user.GetRequest) (*user.UserResponse, error) {
	entity := s.repo.GetByID("user", req.Id)
	if entity == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	userEntity, ok := entity.(*model.User)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast entity to user")
	}

	return &user.UserResponse{
		User: &user.User{
			Id:        userEntity.GetID(),
			Username:  userEntity.GetUsername(),
			Email:     userEntity.GetEmail(),
			CreatedAt: userEntity.GetCreatedAt().Unix(),
			UpdatedAt: userEntity.GetUpdatedAt().Unix(),
		},
	}, nil
}

// UpdateUser обновляет пользователя
func (s *Server) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
	entity := s.repo.GetByID("user", req.Id)
	if entity == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	userEntity, ok := entity.(*model.User)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to cast entity to user")
	}

	userEntity.SetUsername(req.Username)
	userEntity.SetEmail(req.Email)

	s.repo.Save(userEntity)

	return &user.UserResponse{
		User: &user.User{
			Id:        userEntity.GetID(),
			Username:  userEntity.GetUsername(),
			Email:     userEntity.GetEmail(),
			CreatedAt: userEntity.GetCreatedAt().Unix(),
			UpdatedAt: userEntity.GetUpdatedAt().Unix(),
		},
	}, nil
}

// DeleteUser удаляет пользователя
func (s *Server) DeleteUser(ctx context.Context, req *user.GetRequest) (*user.SuccessResponse, error) {
	deleted := s.repo.DeleteByID("user", req.Id)
	if !deleted {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &user.SuccessResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

// ListUsers возвращает список всех пользователей
func (s *Server) ListUsers(ctx context.Context, req *user.Empty) (*user.UsersListResponse, error) {
	entities := s.repo.GetAllByType("user")
	protoUsers := make([]*user.User, len(entities))

	for i, entity := range entities {
		userEntity, ok := entity.(*model.User)
		if !ok {
			log.Printf("Failed to cast entity to user at index %d", i)
			continue
		}

		protoUsers[i] = &user.User{
			Id:        userEntity.GetID(),
			Username:  userEntity.GetUsername(),
			Email:     userEntity.GetEmail(),
			CreatedAt: userEntity.GetCreatedAt().Unix(),
			UpdatedAt: userEntity.GetUpdatedAt().Unix(),
		}
	}

	return &user.UsersListResponse{
		Users: protoUsers,
	}, nil
}
