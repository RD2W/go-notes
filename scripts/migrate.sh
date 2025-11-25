#!/usr/bin/env bash

set -e  # Прекращаем выполнение при ошибке

# Скрипт для запуска миграций базы данных
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
# Пример использования:
#   DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=mypassword DB_NAME=mydb ./migrate.sh
#

# Функция для логирования
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Задаем значения по умолчанию, если переменные окружения не установлены
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"notes_password"}
DB_NAME=${DB_NAME:-"go_notes"}
DB_SSL_MODE=${DB_SSL_MODE:-"disable"}

log "Запуск миграций PostgreSQL..."

# Проверяем, существует ли база данных
log "Проверка существования базы данных..."
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    log "База данных $DB_NAME не существует. Создание базы данных..."
    if PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -O "$DB_USER" "$DB_NAME" 2>/dev/null; then
        log "База данных $DB_NAME создана."
    else
        log "Ошибка при создании базы данных $DB_NAME"
        exit 1
    fi
else
    log "База данных $DB_NAME существует."
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

log "Запуск миграций с использованием строки подключения: $DB_URL"

# Запускаем миграции
if migrate -path ../migrations -database "$DB_URL" -verbose up; then
    log "Миграции успешно завершены!"
else
    log "Ошибка при выполнении миграций"
    exit 1
fi