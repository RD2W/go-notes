package main

import (
	"log"

	grpcclient "github.com/rd2w/go-notes/internal/app/grpc-client"
	"github.com/rd2w/go-notes/internal/config"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию из config_dev.toml: %v", err)
		log.Println("Используем конфигурацию по умолчанию")
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем gRPC-клиент
	client, err := grpcclient.NewGRPCClient(cfg)
	if err != nil {
		log.Fatalf("Ошибка при создании gRPC клиента: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Ошибка при закрытии gRPC клиента: %v", err)
		}
	}()

	// Запускаем тестовую сессию клиента
	client.Run()
}
