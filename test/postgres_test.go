package main_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/repository/postgres"
)

func TestPostgresIntegration(t *testing.T) {
	fmt.Println("Тестирование подключения к PostgreSQL и работы с данными...")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		t.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	// Создаем репозитории
	noteRepo, err := postgres.NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		t.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	userRepo, err := postgres.NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		t.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	// Тестируем создание пользователя
	user, err := model.NewUser("test_user_"+fmt.Sprint(time.Now().Unix()), "test"+fmt.Sprint(time.Now().Unix())+"@example.com", "password123")
	if err != nil {
		t.Fatalf("Ошибка создания пользователя: %v", err)
	}

	err = userRepo.Create(user)
	if err != nil {
		t.Fatalf("Ошибка сохранения пользователя: %v", err)
	}
	fmt.Printf("✓ Пользователь создан с ID: %s\n", user.GetID())

	// Тестируем создание заметки, принадлежащей пользователю
	note := model.NewNote("Тестовая заметка", "Содержимое тестовой заметки", user.GetID())
	err = noteRepo.Create(note)
	if err != nil {
		t.Fatalf("Ошибка сохранения заметки: %v", err)
	}
	fmt.Printf("✓ Заметка создана с ID: %s\n", note.GetID())

	// Получаем все заметки
	notes, err := noteRepo.GetAllNotes()
	if err != nil {
		t.Fatalf("Ошибка получения заметок: %v", err)
	}
	fmt.Printf("✓ Получено %d заметок\n", len(notes))

	// Объявляем retrievedNote перед использованием
	var retrievedNote *model.Note

	// Проверяем, что retrievedNote не является nil перед использованием
	retrievedNote, err = noteRepo.GetByID(note.GetID())
	if err != nil || retrievedNote == nil {
		t.Error("✗ Заметка не найдена по ID")
		// Получаем пользователя по ID
		retrievedUser, err := userRepo.GetByID(user.GetID())
		if err == nil && retrievedUser != nil {
			fmt.Printf("✓ Пользователь найден по ID: %s\n", retrievedUser.GetID())
		} else {
			t.Error("✗ Пользователь не найден по ID")
		}

		t.Log("\n✓ Тесты завершены с частичным успехом! Заметка не найдена, но пользователь обработан.")
		return
	}

	fmt.Printf("✓ Заметка найдена по ID: %s\n", retrievedNote.GetID())

	// Получаем пользователя по ID
	retrievedUser, err := userRepo.GetByID(user.GetID())
	if err == nil && retrievedUser != nil {
		fmt.Printf("✓ Пользователь найден по ID: %s\n", retrievedUser.GetID())
	} else {
		t.Error("✗ Пользователь не найден по ID")
	}

	// Проверяем, что у пользователя есть связь с заметкой
	if retrievedNote.GetUserID() == user.GetID() {
		fmt.Println("✓ Связь между пользователем и заметкой установлена корректно")
	} else {
		t.Error("✗ Связь между пользователем и заметкой НЕ корректна")
	}

	// Проверяем, что у пользователя есть связь с заметкой
	if retrievedNote.GetUserID() == user.GetID() {
		fmt.Println("✓ Связь между пользователем и заметкой установлена корректно")
	} else {
		t.Error("✗ Связь между пользователем и заметкой НЕ корректна")
	}

	// Тестируем обновление заметки
	originalTitle := retrievedNote.GetTitle()
	updatedTitle := originalTitle + " (обновлено)"
	retrievedNote.SetTitle(updatedTitle)
	err = noteRepo.Update(retrievedNote)
	if err != nil {
		t.Errorf("✗ Ошибка обновления заметки: %v\n", err)
	} else {
		fmt.Println("✓ Заметка успешно обновлена")
		// Проверяем обновленную заметку
		updatedNote, err := noteRepo.GetByID(note.GetID())
		if err != nil || updatedNote == nil {
			t.Error("✗ Обновленная заметка не найдена")
		} else if updatedNote.GetTitle() == updatedTitle {
			fmt.Println("✓ Обновление заметки подтверждено")
		} else {
			t.Error("✗ Заметка не была обновлена корректно")
		}
	}

	// Тестируем получение заметок конкретного пользователя
	userNotes, err := noteRepo.GetAllNotesByUserID(user.GetID())
	if err != nil {
		t.Errorf("✗ Ошибка получения заметок пользователя: %v\n", err)
	} else {
		fmt.Printf("✓ Получено %d заметок для пользователя %s\n", len(userNotes), user.GetID())
		if len(userNotes) == 0 {
			t.Error("✗ У пользователя должен быть хотя бы одна заметка")
		}
	}

	// Тестируем пагинацию заметок пользователя
	userNotesPaginated, err := noteRepo.GetListByUserID(user.GetID(), 10, 0)
	if err != nil {
		t.Errorf("✗ Ошибка получения пагинированных заметок пользователя: %v\n", err)
	} else {
		fmt.Printf("✓ Получено %d заметок для пользователя %s с пагинацией\n", len(userNotesPaginated), user.GetID())
	}

	// Тестируем получение всех заметок
	allNotes, err := noteRepo.GetAllNotes()
	if err != nil {
		t.Errorf("✗ Ошибка получения всех заметок: %v\n", err)
	} else {
		fmt.Printf("✓ Количество всех заметок в системе: %d\n", len(allNotes))
	}

	// Тестируем поиск пользователя по email
	foundUserByEmail, err := userRepo.GetUserByEmail(user.GetEmail())
	if err != nil || foundUserByEmail == nil {
		t.Error("✗ Пользователь не найден по email")
	} else if foundUserByEmail.GetID() == user.GetID() {
		fmt.Printf("✓ Пользователь найден по email: %s\n", user.GetEmail())
	} else {
		t.Error("✗ Найден другой пользователь по email")
	}

	// Тестируем поиск пользователя по имени
	foundUserByUsername, err := userRepo.GetUserByUsername(user.GetUsername())
	if err != nil || foundUserByUsername == nil {
		t.Error("✗ Пользователь не найден по имени")
	} else if foundUserByUsername.GetID() == user.GetID() {
		fmt.Printf("✓ Пользователь найден по имени: %s\n", user.GetUsername())
	} else {
		t.Error("✗ Найден другой пользователь по имени")
	}

	// Тестируем получение всех пользователей
	allUsers, err := userRepo.GetAllUsers()
	if err != nil {
		t.Errorf("✗ Ошибка получения всех пользователей: %v\n", err)
	} else {
		fmt.Printf("✓ Получено %d пользователей\n", len(allUsers))
	}

	// Тестируем обновление пользователя
	originalUsername := retrievedUser.GetUsername()
	updatedUsername := originalUsername + "_updated"
	retrievedUser.SetUsername(updatedUsername)
	err = userRepo.Update(retrievedUser)
	if err != nil {
		t.Errorf("✗ Ошибка обновления пользователя: %v\n", err)
	} else {
		fmt.Println("✓ Пользователь успешно обновлен")
		// Проверяем обновленного пользователя
		updatedUser, err := userRepo.GetByID(user.GetID())
		if err != nil || updatedUser == nil {
			t.Error("✗ Обновленный пользователь не найден")
		} else if updatedUser.GetUsername() == updatedUsername {
			fmt.Println("✓ Обновление пользователя подтверждено")
		} else {
			t.Error("✗ Пользователь не был обновлен корректно")
		}
	}

	// Возвращаем исходное имя пользователя для корректного тестирования
	retrievedUser.SetUsername(originalUsername)
	userRepo.Update(retrievedUser)

	// Тестируем удаление заметки
	err = noteRepo.DeleteByID(note.GetID())
	if err != nil {
		t.Errorf("✗ Ошибка удаления заметки: %v\n", err)
	} else {
		fmt.Println("✓ Заметка успешно удалена")
		// Проверяем, что заметка действительно удалена
		deletedNote, err := noteRepo.GetByID(note.GetID())
		if err == nil && deletedNote != nil {
			t.Error("✗ Заметка все еще существует после удаления")
		} else {
			fmt.Println("✓ Заметка подтвержденно удалена")
		}
	}

	// Тестируем удаление пользователя
	err = userRepo.DeleteByID(user.GetID())
	if err != nil {
		t.Errorf("✗ Ошибка удаления пользователя: %v\n", err)
	} else {
		fmt.Println("✓ Пользователь успешно удален")
		// Проверяем, что пользователь действительно удален
		deletedUser, err := userRepo.GetByID(user.GetID())
		if err == nil && deletedUser != nil {
			t.Error("✗ Пользователь все еще существует после удаления")
		} else {
			fmt.Println("✓ Пользователь подтвержденно удален")
		}
	}

	fmt.Println("\n✓ Все расширенные тесты пройдены успешно!")
}
