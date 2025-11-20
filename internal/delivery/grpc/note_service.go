package grpc

import (
	"context"
	"log"

	"github.com/rd2w/go-notes/internal/domain/service"
	notePb "github.com/rd2w/go-notes/pkg/proto/note"
)

// NoteServiceServer реализует gRPC-сервер для сервиса заметок
type NoteServiceServer struct {
	notePb.UnimplementedNotesServiceServer
	noteService service.NoteService
}

// NewNoteServiceServer создает новый экземпляр gRPC-сервера для заметок
func NewNoteServiceServer(noteService service.NoteService) *NoteServiceServer {
	return &NoteServiceServer{
		noteService: noteService,
	}
}

// CreateNote создает новую заметку
func (s *NoteServiceServer) CreateNote(ctx context.Context, req *notePb.CreateNoteRequest) (*notePb.NoteResponse, error) {
	note, err := s.noteService.CreateNote(req.GetTitle(), req.GetContent(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &notePb.NoteResponse{
		Note: &notePb.Note{
			Id:        note.GetID(),
			Title:     note.GetTitle(),
			Content:   note.GetContent(),
			UserId:    note.GetUserID(),
			CreatedAt: note.GetCreatedAt().Unix(),
			UpdatedAt: note.GetUpdatedAt().Unix(),
		},
	}, nil
}

// GetNote возвращает заметку по ID
func (s *NoteServiceServer) GetNote(ctx context.Context, req *notePb.GetRequest) (*notePb.NoteResponse, error) {
	note, err := s.noteService.GetNoteByID(req.Id)
	if err != nil {
		log.Printf("Error getting note: %v", err)
		return nil, err
	}

	return &notePb.NoteResponse{
		Note: &notePb.Note{
			Id:        note.GetID(),
			Title:     note.GetTitle(),
			Content:   note.GetContent(),
			UserId:    note.GetUserID(),
			CreatedAt: note.GetCreatedAt().Unix(),
			UpdatedAt: note.GetUpdatedAt().Unix(),
		},
	}, nil
}

// UpdateNote обновляет заметку
func (s *NoteServiceServer) UpdateNote(ctx context.Context, req *notePb.UpdateNoteRequest) (*notePb.NoteResponse, error) {
	note, err := s.noteService.UpdateNote(req.Id, req.Title, req.Content)
	if err != nil {
		log.Printf("Error updating note: %v", err)
		return nil, err
	}

	return &notePb.NoteResponse{
		Note: &notePb.Note{
			Id:        note.GetID(),
			Title:     note.GetTitle(),
			Content:   note.GetContent(),
			UserId:    note.GetUserID(),
			CreatedAt: note.GetCreatedAt().Unix(),
			UpdatedAt: note.GetUpdatedAt().Unix(),
		},
	}, nil
}

// DeleteNote удаляет заметку
func (s *NoteServiceServer) DeleteNote(ctx context.Context, req *notePb.GetRequest) (*notePb.SuccessResponse, error) {
	err := s.noteService.DeleteNote(req.Id)
	if err != nil {
		log.Printf("Error deleting note: %v", err)
		return nil, err
	}

	return &notePb.SuccessResponse{
		Success: true,
		Message: "note deleted successfully",
	}, nil
}

// ListNotes возвращает список всех заметок
func (s *NoteServiceServer) ListNotes(ctx context.Context, req *notePb.Empty) (*notePb.NotesListResponse, error) {
	notes, err := s.noteService.GetAllNotes()
	if err != nil {
		log.Printf("Error listing notes: %v", err)
		return nil, err
	}

	protoNotes := make([]*notePb.Note, len(notes))
	for i, note := range notes {
		protoNotes[i] = &notePb.Note{
			Id:        note.GetID(),
			Title:     note.GetTitle(),
			Content:   note.GetContent(),
			UserId:    note.GetUserID(),
			CreatedAt: note.GetCreatedAt().Unix(),
			UpdatedAt: note.GetUpdatedAt().Unix(),
		}
	}

	return &notePb.NotesListResponse{
		Notes: protoNotes,
	}, nil
}
