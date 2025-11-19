package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	"github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Добавляем вспомогательную функцию для создания контекста с токеном
func createAuthContext(ctx context.Context, token string) context.Context {
	if token != "" {
		return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
	}
	return ctx
}

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию из config_dev.toml: %v", err)
		log.Println("Используем конфигурацию по умолчанию")
		cfg = config.NewDefaultConfigWithValues()
	}

	// Формируем адрес gRPC сервера
	grpcAddress := "localhost" + cfg.Server.GRPCPort

	// Устанавливаем соединение с gRPC сервером
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения: %v", err)
		}
	}()

	// Создаем клиентов для разных сервисов
	noteClient := note.NewNotesServiceClient(conn)
	userClient := user.NewUserServiceClient(conn)
	authClient := authpb.NewAuthServiceClient(conn)

	// Сначала регистрируем и логиним пользователя для получения токенов
	fmt.Println("=== Регистрация и аутентификация пользователя ===")

	// Создание пользователя
	fmt.Println("\n1. Создание пользователя:")
	createUserResp, err := userClient.CreateUser(context.Background(), &user.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		log.Printf("Ошибка при создании пользователя: %v", err)
	} else {
		fmt.Printf("Создан пользователь: ID=%s, Имя=%s, Email=%s\n", createUserResp.User.Id, createUserResp.User.Username, createUserResp.User.Email)
	}

	// Логинимся для получения токенов
	fmt.Println("\n2. Аутентификация пользователя:")
	loginResp, err := authClient.Login(context.Background(), &authpb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		log.Fatalf("Ошибка при аутентификации: %v", err)
	}
	fmt.Printf("Успешная аутентификация. Access токен: %s\n", loginResp.AccessToken)

	// Создаем контекст с токеном для аутентифицированных запросов
	authCtx := createAuthContext(context.Background(), loginResp.AccessToken)

	// Тестирование операций с заметками с аутентификацией
	fmt.Println("\n=== Тестирование операций с заметками (с аутентификацией) ===")

	// Создание заметки
	fmt.Println("\n3. Создание заметки:")
	createNoteResp, err := noteClient.CreateNote(authCtx, &note.CreateNoteRequest{
		Title:   "Тестовая заметка",
		Content: "Это содержимое тестовой заметки",
	})
	if err != nil {
		log.Printf("Ошибка при создании заметки: %v", err)
	} else {
		fmt.Printf("Создана заметка: ID=%s, Заголовок=%s\n", createNoteResp.Note.Id, createNoteResp.Note.Title)
	}

	// Получение списка заметок
	fmt.Println("\n4. Получение списка заметок:")
	listNotesResp, err := noteClient.ListNotes(authCtx, &note.Empty{})
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
		getNoteResp, err := noteClient.GetNote(authCtx, &note.GetRequest{Id: firstNote.Id})
		if err != nil {
			log.Printf("Ошибка при получении заметки: %v", err)
		} else {
			fmt.Printf("Получена заметка: ID=%s, Заголовок=%s, Содержимое=%s\n",
				getNoteResp.Note.Id, getNoteResp.Note.Title, getNoteResp.Note.Content)
		}

		fmt.Printf("\n6. Обновление заметки (%s):\n", firstNote.Id)
		updateNoteResp, err := noteClient.UpdateNote(authCtx, &note.UpdateNoteRequest{
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
	listUsersResp, err := userClient.ListUsers(authCtx, &user.Empty{})
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
		getUserResp, err := userClient.GetUser(authCtx, &user.GetRequest{Id: firstUser.Id})
		if err != nil {
			log.Printf("Ошибка при получении пользователя: %v", err)
		} else {
			fmt.Printf("Получен пользователь: ID=%s, Имя=%s, Email=%s\n",
				getUserResp.User.Id, getUserResp.User.Username, getUserResp.User.Email)
		}

		fmt.Printf("\n9. Обновление пользователя (%s):\n", firstUser.Id)
		updateUserResp, err := userClient.UpdateUser(authCtx, &user.UpdateUserRequest{
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
		deleteNoteResp, err := noteClient.DeleteNote(authCtx, &note.GetRequest{Id: createNoteResp.Note.Id})
		if err != nil {
			log.Printf("Ошибка при удалении заметки: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteNoteResp.Success, deleteNoteResp.Message)
		}
	}

	// Удаление последнего созданного пользователя
	if createUserResp != nil {
		fmt.Printf("\n11. Удаление пользователя (%s):\n", createUserResp.User.Id)
		deleteUserResp, err := userClient.DeleteUser(authCtx, &user.GetRequest{Id: createUserResp.User.Id})
		if err != nil {
			log.Printf("Ошибка при удалении пользователя: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteUserResp.Success, deleteUserResp.Message)
		}
	}

	// Логаут
	fmt.Println("\n12. Выход из системы:")
	logoutResp, err := authClient.Logout(authCtx, &authpb.LogoutRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		log.Printf("Ошибка при выходе: %v", err)
	} else {
		fmt.Printf("Результат выхода: %t, Сообщение: %s\n", logoutResp.Success, logoutResp.Message)
	}

	fmt.Println("\nТестирование gRPC клиента с аутентификацией завершено.")
}
