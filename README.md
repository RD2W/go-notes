# Go Notes

Репозиторий для изучения языка Go, демонстрирующий создание веб-сервера и gRPC-сервера с возможностью управления заметками и пользователями.

[![Release](https://img.shields.io/github/v/release/RD2W/go-notes.svg?label=Release)](https://github.com/RD2W/go-notes/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Test Coverage](https://img.shields.io/badge/Coverage-Test%20Coverage-1abc9c.svg)](https://golang.org)
![Tests](https://github.com/RD2W/go-notes/actions/workflows/branch_merges.yml/badge.svg)

## Особенности проекта

- **Веб-API**: RESTful API с использованием фреймворка Gin
- **gRPC-сервер**: Реализация gRPC-сервисов для заметок и пользователей
- **Аутентификация**: JWT-токены для защиты маршрутов
- **Хранение данных**: Поддержка различных хранилищ (PostgreSQL для основных данных и Redis для токенов)
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
├── config/                 # Конфигурационные файлы
├── docs/                   # Документация Swagger
├── internal/               # Внутренний код приложения
│   ├── app/                # Входные точки приложения
│   ├── auth/               # Аутентификация и токены
│   ├── config/             # Управление конфигурацией
│   ├── database/           # Подключения к базам данных
│   ├── delivery/           # Контроллеры/обработчики HTTP и gRPC
│   ├── domain/             # Бизнес-логика и модели
│   │   ├── model/          # Определения структур данных
│   │   └── repository/     # Интерфейсы репозиториев
│   ├── middleware/         # HTTP-мидлвары (например, аутентификация)
│   ├── repository/         # Реализации репозиториев (PostgreSQL, Redis)
│   ├── service/            # Бизнес-сервисы
│   └── util/               # Вспомогательные утилиты
├── migrations/             # SQL-скрипты миграций
├── pkg/                    # Публичные пакеты (сгенерированный protobuf-код)
├── scripts/                # Скрипты для генерации кода
├── test/                   # Тесты
├── docker-compose.yml      # Конфигурация Docker Compose
├── go.mod                  # Зависимости Go
├── go.sum                  # Чек-суммы зависимостей
├── LICENSE                 # Лицензия
├── Makefile                # Сборочные команды
└── README.md               # Документация проекта
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
- PostgreSQL для хранения пользователей и заметок
- Redis для хранения отозванных токенов и кэширования

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

Сервер будет доступен по адресу `localhost:5051`.

### Запуск gRPC-клиента

```bash
go run cmd/grpc-client/main.go
```

Клиент выполнит тестовые операции с gRPC-сервером.

### Запуск с Docker

```bash
# Запуск PostgreSQL и Redis с помощью Docker Compose
make docker-up

# Запуск миграций базы данных
make migrate

# Откат всех миграций базы данных
make cleanup-db

# Полный перезапуск с Docker и миграции
make setup-db
```

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
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

#### Выход пользователя (отзыв refresh токена)
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {jwt_token}" \
  -d '{"refresh_token": "{refresh_token}"}'
```

#### Обновление токенов
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "{refresh_token}"}'
```

#### Проверка валидности токена
```bash
curl -X POST http://localhost:8080/api/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"token": "{jwt_token}"}'
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
- `make docker-down-v` - остановка контейнеров и удаление volumes с БД
- `make docker-up` - запуск сервисов с Docker Compose
- `make migrate` - запуск миграций базы данных
- `make cleanup-db` - откат всех миграций базы данных
- `make setup-db` - полный перезапуск с Docker и запуск миграций
- `make help` - список всех целей

## Переменные окружения

Приложение поддерживает настройку через переменные окружения. Ниже приведены доступные переменные:

### Общие настройки
- `ENV` - окружение (development, production, staging) (по умолчанию: development)
- `LOG_LEVEL` - уровень логирования (по умолчанию: info)

### Сервер
- `SERVER_PORT` - порт для HTTP сервера (по умолчанию: :8080)
- `GRPC_PORT` - порт для gRPC сервера (по умолчанию: :5051)
- `SERVER_READ_TIMEOUT` - таймаут чтения запроса в секундах (по умолчанию: 15)
- `SERVER_WRITE_TIMEOUT` - таймаут записи ответа в секундах (по умолчанию: 15)

### Аутентификация
- `AUTH_ENABLE_HTTPS` - включить HTTPS (по умолчанию: false)

### PostgreSQL
- `POSTGRES_HOST` - хост PostgreSQL (по умолчанию: localhost)
- `POSTGRES_PORT` - порт PostgreSQL (по умолчанию: 5432)
- `POSTGRES_NAME` - имя базы данных (по умолчанию: go_notes)
- `POSTGRES_USER` - имя пользователя базы данных (по умолчанию: postgres)
- `POSTGRES_PASSWORD` - пароль базы данных
- `POSTGRES_SSL_MODE` - режим SSL для PostgreSQL (по умолчанию: disable)
- `POSTGRES_POOL_SIZE` - размер пула подключений (по умолчанию: 10)
- `POSTGRES_PARAMETERS` - дополнительные параметры подключения

### Redis
- `REDIS_HOST` - хост Redis (по умолчанию: localhost)
- `REDIS_PORT` - порт Redis (по умолчанию: 6379)
- `REDIS_PASSWORD` - пароль Redis
- `REDIS_DB` - номер базы данных Redis (по умолчанию: 0)
- `REDIS_POOL_SIZE` - размер пула подключений (по умолчанию: 10)
- `REDIS_URL` - альтернативный способ указания подключения

### JWT
- `JWT_SECRET_KEY` - секретный ключ для подписи JWT токенов (по умолчанию: my_secret_key)
- `JWT_ALGORITHM` - алгоритм подписи токена (по умолчанию: HS256)
- `BCRYPT_COST` - стоимость хеширования паролей (по умолчанию: 10)
- `ACCESS_TOKEN_TTL` - время жизни access токена (по умолчанию: 15m)
- `REFRESH_TOKEN_TTL` - время жизни refresh токена (по умолчанию: 168h)

### Refresh токены
- `REFRESH_SECRET_KEY` - секретный ключ для подписи Refresh токенов (по умолчанию: refresh_secret_key)
- `REFRESH_REVOCATION_ENABLED` - включено ли отслеживание отозванных токенов (по умолчанию: true)

### Репозиторий
- `REPO_TYPE` - тип репозитория (postgres, redis) (по умолчанию: postgres)
- `REPO_PATH` - путь к файлу/директории для хранения данных (по умолчанию: ./data)

### Безопасность
- `PASSWORD_MIN_LENGTH` - минимальная длина пароля (по умолчанию: 8)
- `MAX_LOGIN_ATTEMPTS` - максимальное количество попыток входа (по умолчанию: 5)
- `LOGIN_BLOCK_TIME` - время блокировки после неудачных попыток (по умолчанию: 30m)
- `BCRYPT_COST_SEC` - стоимость хеширования паролей (по умолчанию: 10)

### Завершение работы
- `SHUTDOWN_TIMEOUT` - таймаут завершения работы (по умолчанию: 25s)
- `SHUTDOWN_WAIT` - время ожидания перед завершением (по умолчанию: 3s)
