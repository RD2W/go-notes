#!/usr/bin/env bash

set -e  # Прекращаем выполнение при ошибке

# Скрипт для запуска миграций базы данных

echo "Запуск миграций PostgreSQL..."

# Устанавливаем миграционный инструмент, если он не установлен
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Запускаем миграции
migrate -path ../migrations -database 'postgres://postgres:notes_password@localhost:5432/go_notes?sslmode=disable' -verbose up

echo "Миграции завершены!"