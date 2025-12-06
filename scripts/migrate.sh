#!/usr/bin/env bash

set -e  # Прекращаем выполнение при ошибке

# Скрипт для запуска миграций базы данных или отката миграций
# Использует переменные окружения для подключения к базе данных
#
# Переменные окружения:
#   DB_HOST - хост базы данных (по умолчанию: localhost)
#   DB_PORT - порт базы данных (по умолчанию: 5432)
#   DB_USER - пользователь базы данных (по умолчанию: postgres)
#   DB_PASSWORD - пароль базы данных (по умолчанию: notes_password)
#   DB_NAME - имя базы данных (по умолчанию: go_notes)
#   DB_SSL_MODE - режим SSL (по умолчанию: disable)
#
# Параметры:
#   up - применить миграции (по умолчанию)
#   down - откатить все миграции
#
# Пример использования:
#   ./migrate.sh up
#   ./migrate.sh down
#   DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=mypassword DB_NAME=mydb ./migrate.sh up
#

# Функция для логирования
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Функция для отображения справки
show_help() {
    echo "Использование: $0 [up|down]"
    echo "  up   - применить миграции (по умолчанию)"
    echo " down - откатить все миграции"
    echo ""
    echo "Переменные окружения:"
    echo "  DB_HOST     - хост базы данных (по умолчанию: localhost)"
    echo "  DB_PORT     - порт базы данных (по умолчанию: 5432)"
    echo "  DB_USER     - пользователь базы данных (по умолчанию: postgres)"
    echo "  DB_PASSWORD - пароль базы данных (по умолчанию: notes_password)"
    echo "  DB_NAME     - имя базы данных (по умолчанию: go_notes)"
    echo "  DB_SSL_MODE - режим SSL (по умолчанию: disable)"
}

# Задаем значения по умолчанию, если переменные окружения не установлены
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"notes_password"}
DB_NAME=${DB_NAME:-"go_notes"}
DB_SSL_MODE=${DB_SSL_MODE:-"disable"}

# Проверяем аргумент командной строки
ACTION=${1:-"up"}

# Проверяем, является ли аргумент допустимым
if [[ "$ACTION" != "up" && "$ACTION" != "down" && "$ACTION" != "-h" && "$ACTION" != "--help" ]]; then
    log "Недопустимый аргумент: $ACTION"
    show_help
    exit 1
fi

# Показываем справку, если запрошено
if [[ "$ACTION" == "-h" || "$ACTION" == "--help" ]]; then
    show_help
    exit 0
fi

# Проверяем, существует ли база данных через Docker
log "Проверка существования базы данных..."
if docker exec go-notes-postgres psql -U "$DB_USER" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" 2>/dev/null | grep -q 1; then
    log "База данных $DB_NAME существует."
else
    if [[ "$ACTION" == "down" ]]; then
        log "База данных $DB_NAME не существует. Выход."
        exit 0
    else
        log "База данных $DB_NAME не существует. Создание базы данных..."
        if docker exec go-notes-postgres psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null; then
            log "База данных $DB_NAME создана."
        else
            log "Ошибка при создании базы данных $DB_NAME"
            exit 1
        fi
    fi
fi

# Устанавливаем миграционный инструмент, если он не установлен
log "Проверка и установка миграционного инструмента..."
if ! command -v migrate &> /dev/null; then
    log "Миграционный инструмент не найден, устанавливаем..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    if [ $? -ne 0 ]; then
        log "Ошибка при установке миграционного инструмента"
        exit 1
    fi
else
    log "Миграционный инструмент уже установлен."
fi

# Формируем строку подключения к базе данных
DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE"

if [[ "$ACTION" == "up" ]]; then
    log "Запуск миграций PostgreSQL..."
    log "Запуск миграций с использованием строки подключения: $DB_URL"
    
    # Запускаем миграции
    if migrate -path ./migrations -database "$DB_URL" -verbose up; then
        log "Миграции успешно завершены!"
    else
        log "Ошибка при выполнении миграций"
        exit 1
    fi
else
    log "Запуск очистки базы данных PostgreSQL..."
    log "Запуск отката миграций с использованием строки подключения: $DB_URL"
    
    # Запускаем откат миграций
    if migrate -path ./migrations -database "$DB_URL" -verbose down -all; then
        log "Откат миграций успешно завершен!"
    else
        log "Ошибка при выполнении отката миграций"
        exit 1
    fi
fi