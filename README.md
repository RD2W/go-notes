# Go Notes

Репозиторий для изучения языка Go, демонстрирующий создание веб-сервера и gRPC-сервера с возможностью управления заметками и пользователями.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Особенности проекта

- **Веб-API**: RESTful API с использованием фреймворка Gin
- **gRPC-сервер**: Реализация gRPC-сервисов для заметок и пользователей
- **Аутентификация**: JWT-токены для защиты маршрутов
- **Хранение данных**: Поддержка различных хранилищ (в памяти и в JSON-файлах)
- **Документация API**: Swagger UI для веб-API
- **Protocol Buffers**: Для определения gRPC-сервисов
- **Тестирование**: Модульные тесты

## Структура проекта

```
go-notes/
├── api/                    # Определения API (protobuf)
├── cmd/                    # Основные приложения
│   ├── grpc-client/        # Клиент gRPC
│   ├── grpc-server/        # Сервер gRPC
│   └── web-server/         # Веб-сервер (REST API)
├── docs/                   # Документация Swagger
├── internal/               # Внутренний код приложения
│   ├── grpc/               # Реализация gRPC-сервера
│   ├── handler/            # Обработчики HTTP-запросов
│   ├── middleware/         # HTTP-мидлвары (например, аутентификация)
│   ├── model/              # Определения структур данных
│   ├── repository/         # Интерфейсы и фабрики репозиториев
│   └── util/               # Вспомогательные утилиты
├── pkg/                    # Публичные пакеты (сгенерированный protobuf-код)
├── scripts/                # Скрипты для генерации кода
└── Makefile                # Сборочные команды
```

## Функциональность

### Веб-сервер (REST API)
- Аутентификация пользователей через JWT
- CRUD-операции для заметок и пользователей
- Swagger UI доступен по адресу `/swagger/index.html`
- Защищенные маршруты для изменения данных
- Открытые маршруты для чтения данных

### gRPC-сервер
- Сервис для управления заметками
- Сервис для управления пользователями
- Поддержка всех CRUD-операций через gRPC

### Хранение данных
- RAM-хранилище для временных данных
- JSON-хранилище для сохранения данных между запусками

## Запуск приложения

### Предварительные требования
- Go 1.25.4 или выше
- protoc (компилятор Protocol Buffers)
- make

### Установка зависимостей

```bash
# Установка зависимостей для protobuf
make proto-deps

# Генерация protobuf-кода
make proto

# Установка зависимостей для Swagger
make swag-deps

# Генерация документации Swagger
make swag
```

### Запуск веб-сервера

```bash
go run cmd/web-server/main.go
```

Сервер будет доступен по адресу `http://localhost:8080`, Swagger UI по адресу `http://localhost:8080/swagger/index.html`.

### Запуск gRPC-сервера

```bash
go run cmd/grpc-server/main.go
```

Сервер будет доступен по адресу `localhost:50051`.

### Запуск gRPC-клиента

```bash
go run cmd/grpc-client/main.go
```

Клиент выполнит тестовые операции с gRPC-сервером.

## Примеры использования API

### Работа с пользователями

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "password123"}'
```

#### Аутентификация пользователя (получение JWT-токена)
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
 -d '{"username": "testuser", "password": "password123"}'
```

После успешной аутентификации вы получите JWT-токен. При использовании токена в других запросах не включайте фигурные скобки `{}` - они используются только для обозначения плейсхолдера в примерах.

#### Получение всех пользователей
```bash
curl -X GET http://localhost:8080/api/users
```

#### Получение пользователя по ID
```bash
curl -X GET http://localhost:8080/api/users/{user_id}
```

#### Обновление пользователя (требует JWT-токен)
```bash
curl -X PUT http://localhost:8080/api/users/{user_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {jwt_token}" \
  -d '{"username": "updateduser", "email": "updated@example.com", "password": ""}'
```

#### Удаление пользователя (требует JWT-токен)
```bash
curl -X DELETE http://localhost:8080/api/users/{user_id} \
  -H "Authorization: Bearer {jwt_token}"
```

### Работа с заметками

#### Создание заметки (требует JWT-токен)
```bash
curl -X POST http://localhost:8080/api/notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {jwt_token}" \
  -d '{"title": "Моя заметка", "content": "Содержимое заметки"}'
```

#### Получение всех заметок (открытый маршрут)
```bash
curl -X GET http://localhost:8080/api/notes
```

#### Получение заметки по ID (требует JWT-токен)
```bash
curl -X GET http://localhost:8080/api/notes/{note_id} \
  -H "Authorization: Bearer {jwt_token}"
```

#### Обновление заметки (требует JWT-токен)
```bash
curl -X PUT http://localhost:8080/api/notes/{note_id} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {jwt_token}" \
  -d '{"id": "{note_id}", "title": "Обновленная заметка", "content": "Обновленное содержимое"}'
```

#### Удаление заметки (требует JWT-токен)
```bash
curl -X DELETE http://localhost:8080/api/notes/{note_id} \
  -H "Authorization: Bearer {jwt_token}"
```

## Используемые технологии

- [Gin](https://github.com/gin-gonic/gin) - веб-фреймворк
- [gRPC](https://grpc.io/) - фреймворк для RPC
- [Protocol Buffers](https://developers.google.com/protocol-buffers) - язык описания схемы данных
- [JWT](https://jwt.io/) - токены для аутентификации
- [Swaggo](https://github.com/swaggo/swag) - генерация документации Swagger
- [Testify](https://github.com/stretchr/testify) - библиотека для тестирования

## Make-цели

- `make proto` - генерация protobuf-кода
- `make proto-deps` - установка зависимостей protobuf
- `make swag` - генерация документации Swagger
- `make swag-deps` - установка зависимостей Swagger
- `make test` - запуск тестов
- `make build` - сборка приложения
- `make help` - список всех целей
