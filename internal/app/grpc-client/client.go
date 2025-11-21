package grpcclient

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	"github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCClient структура gRPC-клиента
type GRPCClient struct {
	config     *config.Config
	conn       *grpc.ClientConn
	noteClient note.NotesServiceClient
	userClient user.UserServiceClient
	authClient authpb.AuthServiceClient
}

// NewGRPCClient создает новый экземпляр gRPC-клиента
func NewGRPCClient(cfg *config.Config) (*GRPCClient, error) {
	// Формируем адрес gRPC сервера
	grpcAddress := "localhost" + cfg.Server.GRPCPort

	// Устанавливаем соединение с gRPC сервером
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
	}

	// Создаем клиентов для разных сервисов
	noteClient := note.NewNotesServiceClient(conn)
	userClient := user.NewUserServiceClient(conn)
	authClient := authpb.NewAuthServiceClient(conn)

	client := &GRPCClient{
		config:     cfg,
		conn:       conn,
		noteClient: noteClient,
		userClient: userClient,
		authClient: authClient,
	}

	return client, nil
}

// Close закрывает соединение с gRPC сервером
func (gc *GRPCClient) Close() error {
	return gc.conn.Close()
}

// GetNoteClient возвращает клиент для работы с заметками
func (gc *GRPCClient) GetNoteClient() note.NotesServiceClient {
	return gc.noteClient
}

// GetUserClient возвращает клиент для работы с пользователями
func (gc *GRPCClient) GetUserClient() user.UserServiceClient {
	return gc.userClient
}

// GetAuthClient возвращает клиент для работы с аутентификацией
func (gc *GRPCClient) GetAuthClient() authpb.AuthServiceClient {
	return gc.authClient
}

// GetConfig возвращает конфигурацию клиента
func (gc *GRPCClient) GetConfig() *config.Config {
	return gc.config
}

// Run запускает тестовую сессию клиента
func (gc *GRPCClient) Run() {
	// Сначала регистрируем и логиним пользователя для получения токенов
	fmt.Println("=== Регистрация и аутентификация пользователя ===")

	// Создание пользователя
	fmt.Println("\n1. Создание пользователя:")

	// Генерируем уникальное имя пользователя и email
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	email := fmt.Sprintf("test_%d@example.com", time.Now().Unix())
	password := "password123"

	// Создаем пользователя с уникальными данными
	createUserResp, err := gc.userClient.CreateUser(context.Background(), &user.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
	})

	if err != nil {
		log.Printf("Ошибка при создании пользователя: %v", err)
		return
	} else {
		fmt.Printf("Создан пользователь: ID=%s, Имя=%s, Email=%s\n", createUserResp.User.Id, createUserResp.User.Username, createUserResp.User.Email)
	}

	// Логинимся для получения токенов
	fmt.Println("\n2. Аутентификация пользователя:")
	loginResp, err := gc.authClient.Login(context.Background(), &authpb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Printf("Ошибка при аутентификации: %v", err)
		// Если аутентификация не удалась, завершаем выполнение
		return
	}
	fmt.Printf("Успешная аутентификация. Access токен: %s\n", loginResp.AccessToken)

	// Создаем контекст с токеном для аутентифицированных запросов
	authCtx := gc.createAuthContext(context.Background(), loginResp.AccessToken)

	// Тестирование операций с заметками с аутентификацией
	fmt.Println("\n=== Тестирование операций с заметками (с аутентификацией) ===")

	// Создание заметки
	fmt.Println("\n3. Создание заметки:")
	createNoteResp, err := gc.noteClient.CreateNote(authCtx, &note.CreateNoteRequest{
		Title:   "Тестовая заметка",
		Content: "Это содержимое тестовой заметки",
		UserId:  createUserResp.User.Id,
	})
	if err != nil {
		log.Printf("Ошибка при создании заметки: %v", err)
	} else {
		fmt.Printf("Создана заметка: ID=%s, Заголовок=%s\n", createNoteResp.Note.Id, createNoteResp.Note.Title)
	}

	// Получение списка заметок
	fmt.Println("\n4. Получение списка заметок:")
	listNotesResp, err := gc.noteClient.ListNotes(authCtx, &note.Empty{})
	if err != nil {
		log.Printf("Ошибка при получении списка заметок: %v", err)
	} else {
		fmt.Printf("Найдено %d заметок:\n", len(listNotesResp.Notes))
		for _, noteItem := range listNotesResp.Notes {
			fmt.Printf(" - ID: %s, Заголовок: %s\n", noteItem.Id, noteItem.Title)
		}
	}

	// Если есть хотя бы одна заметка, получаем её по ID и обновляем
	if len(listNotesResp.Notes) > 0 {
		firstNote := listNotesResp.Notes[0]
		fmt.Printf("\n5. Получение заметки по ID (%s):\n", firstNote.Id)
		getNoteResp, err := gc.noteClient.GetNote(authCtx, &note.GetRequest{Id: firstNote.Id})
		if err != nil {
			log.Printf("Ошибка при получении заметки: %v", err)
		} else {
			fmt.Printf("Получена заметка: ID=%s, Заголовок=%s, Содержимое=%s\n",
				getNoteResp.Note.Id, getNoteResp.Note.Title, getNoteResp.Note.Content)
		}

		fmt.Printf("\n6. Обновление заметки (%s):\n", firstNote.Id)
		updateNoteResp, err := gc.noteClient.UpdateNote(authCtx, &note.UpdateNoteRequest{
			Id:      firstNote.Id,
			Title:   "Обновленная тестовая заметка",
			Content: "Это обновленное содержимое тестовой заметки",
		})
		if err != nil {
			log.Printf("Ошибка при обновлении заметки: %v", err)
		} else {
			fmt.Printf("Обновлена заметка: ID=%s, Заголовок=%s\n", updateNoteResp.Note.Id, updateNoteResp.Note.Title)
		}
	}

	// Тестирование операций с пользователями с аутентификацией
	fmt.Println("\n=== Тестирование операций с пользователями (с аутентификацией) ===")

	// Получение списка пользователей
	fmt.Println("\n7. Получение списка пользователей:")
	listUsersResp, err := gc.userClient.ListUsers(authCtx, &user.Empty{})
	if err != nil {
		log.Printf("Ошибка при получении списка пользователей: %v", err)
	} else {
		fmt.Printf("Найдено %d пользователей:\n", len(listUsersResp.Users))
		for _, userItem := range listUsersResp.Users {
			fmt.Printf(" - ID: %s, Имя: %s, Email: %s\n", userItem.Id, userItem.Username, userItem.Email)
		}
	}

	// Если есть хотя бы один пользователь, получаем его по ID и обновляем
	if len(listUsersResp.Users) > 0 {
		firstUser := listUsersResp.Users[0]
		fmt.Printf("\n8. Получение пользователя по ID (%s):\n", firstUser.Id)
		getUserResp, err := gc.userClient.GetUser(authCtx, &user.GetRequest{Id: firstUser.Id})
		if err != nil {
			log.Printf("Ошибка при получении пользователя: %v", err)
		} else {
			fmt.Printf("Получен пользователь: ID=%s, Имя=%s, Email=%s\n",
				getUserResp.User.Id, getUserResp.User.Username, getUserResp.User.Email)
		}

		fmt.Printf("\n9. Обновление пользователя (%s):\n", firstUser.Id)
		updateUserResp, err := gc.userClient.UpdateUser(authCtx, &user.UpdateUserRequest{
			Id:       firstUser.Id,
			Username: "updateduser",
			Email:    "updated@example.com",
		})
		if err != nil {
			log.Printf("Ошибка при обновлении пользователя: %v", err)
		} else {
			fmt.Printf("Обновлен пользователь: ID=%s, Имя=%s, Email=%s\n", updateUserResp.User.Id, updateUserResp.User.Username, updateUserResp.User.Email)
		}
	}

	// Небольшая задержка перед удалением
	time.Sleep(1 * time.Second)

	// Удаление последней созданной заметки
	if createNoteResp != nil {
		fmt.Printf("\n10. Удаление заметки (%s):\n", createNoteResp.Note.Id)
		deleteNoteResp, err := gc.noteClient.DeleteNote(authCtx, &note.GetRequest{Id: createNoteResp.Note.Id})
		if err != nil {
			log.Printf("Ошибка при удалении заметки: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteNoteResp.Success, deleteNoteResp.Message)
		}
	}

	// Удаление последнего созданного пользователя
	if createUserResp != nil {
		fmt.Printf("\n11. Удаление пользователя (%s):\n", createUserResp.User.Id)
		deleteUserResp, err := gc.userClient.DeleteUser(authCtx, &user.GetRequest{Id: createUserResp.User.Id})
		if err != nil {
			log.Printf("Ошибка при удалении пользователя: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteUserResp.Success, deleteUserResp.Message)
		}
	}

	// Логаут
	fmt.Println("\n12. Выход из системы:")
	logoutResp, err := gc.authClient.Logout(authCtx, &authpb.LogoutRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		log.Printf("Ошибка при выходе: %v", err)
	} else {
		fmt.Printf("Результат выхода: %t, Сообщение: %s\n", logoutResp.Success, logoutResp.Message)
	}

	// Проверяем, что токен действительно отозван - пытаемся создать заметку после разлогинивания
	fmt.Println("\n13. Проверка использования отозванного токена:")
	// Создаем новый контекст с тем же токеном
	invalidCtx := gc.createAuthContext(context.Background(), loginResp.AccessToken)
	_, err = gc.noteClient.CreateNote(invalidCtx, &note.CreateNoteRequest{
		Title:   "Тестовая заметка после логаута",
		Content: "Эта заметка не должна быть создана",
		UserId:  createUserResp.User.Id,
	})
	if err != nil {
		fmt.Printf("Токен успешно отозван, ошибка при создании заметки: %v\n", err)
		// Проверяем, содержит ли ошибка сообщение об отозванном токене
		if strings.Contains(err.Error(), "токен был отозван") || strings.Contains(err.Error(), "token has been revoked") {
			fmt.Println("Токен был корректно отозван и не может быть использован")
		} else {
			fmt.Println("Ошибка связана с чем-то другим, не с отозванием токена")
		}
	} else {
		fmt.Println("Токен не был отозван, заметка создана")
	}

	fmt.Println("\nТестирование gRPC клиента с аутентификацией завершено.")
}

// createAuthContext создает контекст с токеном аутентификации
func (gc *GRPCClient) createAuthContext(ctx context.Context, token string) context.Context {
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
	}
	return ctx
}
