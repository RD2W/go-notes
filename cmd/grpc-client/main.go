package main

import (
	"context"
	"fmt"
	"log"
	"time"

	note "github.com/rd2w/go-notes/pkg/proto/note"
	user "github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
)

func main() {
	// Устанавливаем соединение с gRPC сервером
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к gRPC серверу: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения: %v", err)
		}
	}()

	client := note.NewNotesServiceClient(conn)

	// Тестирование операций с заметками
	fmt.Println("=== Тестирование операций с заметками ===")

	// Создание заметки
	fmt.Println("\n1. Создание заметки:")
	createNoteResp, err := client.CreateNote(context.Background(), &note.CreateNoteRequest{
		Title:   "Тестовая заметка",
		Content: "Это содержимое тестовой заметки",
	})
	if err != nil {
		log.Printf("Ошибка при создании заметки: %v", err)
	} else {
		fmt.Printf("Создана заметка: ID=%s, Заголовок=%s\n", createNoteResp.Note.Id, createNoteResp.Note.Title)
	}

	// Получение списка заметок
	fmt.Println("\n2. Получение списка заметок:")
	listNotesResp, err := client.ListNotes(context.Background(), &note.Empty{})
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
		fmt.Printf("\n3. Получение заметки по ID (%s):\n", firstNote.Id)
		getNoteResp, err := client.GetNote(context.Background(), &note.GetRequest{Id: firstNote.Id})
		if err != nil {
			log.Printf("Ошибка при получении заметки: %v", err)
		} else {
			fmt.Printf("Получена заметка: ID=%s, Заголовок=%s, Содержимое=%s\n",
				getNoteResp.Note.Id, getNoteResp.Note.Title, getNoteResp.Note.Content)
		}

		fmt.Printf("\n4. Обновление заметки (%s):\n", firstNote.Id)
		updateNoteResp, err := client.UpdateNote(context.Background(), &note.UpdateNoteRequest{
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

	// Тестирование операций с пользователями
	fmt.Println("\n=== Тестирование операций с пользователями ===")

	// Создание пользователя
	fmt.Println("\n1. Создание пользователя:")
	userClient := user.NewUserServiceClient(conn)
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

	// Получение списка пользователей
	fmt.Println("\n2. Получение списка пользователей:")
	listUsersResp, err := userClient.ListUsers(context.Background(), &user.Empty{})
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
		fmt.Printf("\n3. Получение пользователя по ID (%s):\n", firstUser.Id)
		getUserResp, err := userClient.GetUser(context.Background(), &user.GetRequest{Id: firstUser.Id})
		if err != nil {
			log.Printf("Ошибка при получении пользователя: %v", err)
		} else {
			fmt.Printf("Получен пользователь: ID=%s, Имя=%s, Email=%s\n",
				getUserResp.User.Id, getUserResp.User.Username, getUserResp.User.Email)
		}

		fmt.Printf("\n4. Обновление пользователя (%s):\n", firstUser.Id)
		updateUserResp, err := userClient.UpdateUser(context.Background(), &user.UpdateUserRequest{
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
		fmt.Printf("\n5. Удаление заметки (%s):\n", createNoteResp.Note.Id)
		deleteNoteResp, err := client.DeleteNote(context.Background(), &note.GetRequest{Id: createNoteResp.Note.Id})
		if err != nil {
			log.Printf("Ошибка при удалении заметки: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteNoteResp.Success, deleteNoteResp.Message)
		}
	}

	// Удаление последнего созданного пользователя
	if createUserResp != nil {
		fmt.Printf("\n6. Удаление пользователя (%s):\n", createUserResp.User.Id)
		deleteUserResp, err := userClient.DeleteUser(context.Background(), &user.GetRequest{Id: createUserResp.User.Id})
		if err != nil {
			log.Printf("Ошибка при удалении пользователя: %v", err)
		} else {
			fmt.Printf("Результат удаления: %t, Сообщение: %s\n", deleteUserResp.Success, deleteUserResp.Message)
		}
	}

	fmt.Println("\nТестирование gRPC клиента завершено.")
}
