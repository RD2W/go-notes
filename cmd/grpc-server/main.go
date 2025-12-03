package main

import (
	"log"

	grpcserver "github.com/rd2w/go-notes/internal/app/grpc-server"
	"github.com/rd2w/go-notes/internal/config"
)

// Переменные для версионирования, заполняются при сборке
var (
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"
)

func main() {
	// Выводим информацию о версии
	log.Printf("Запуск gRPC-сервера. Версия: %s, Коммит: %s, Дата сборки: %s", version, commit, date)

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config_dev.toml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем и запускаем gRPC-сервер
	grpcServer := grpcserver.NewGRPCServer(cfg)
	if err := grpcServer.Run(); err != nil {
		log.Fatalf("Ошибка при запуске gRPC сервера: %v", err)
	}
}
