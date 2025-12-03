package main

import (
	"log"

	webserver "github.com/rd2w/go-notes/internal/app/web-server"
	"github.com/rd2w/go-notes/internal/config"

	_ "github.com/rd2w/go-notes/docs"
)

// Переменные для версионирования, заполняются при сборке
var (
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"
)

// @title Go Notes API
// @version 1.0
// @description API для управления заметками и пользователями
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme
func main() {
	// Выводим информацию о версии
	log.Printf("Запуск веб-сервера. Версия: %s, Коммит: %s, Дата сборки: %s", version, commit, date)

	// Загружаем конфигурацию из файла
	cfg, err := config.LoadConfig("./config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию из config_dev.toml: %v", err)
		log.Println("Используем конфигурацию по умолчанию")
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем и запускаем веб-сервер
	webServer := webserver.NewWebServer(cfg)
	if err := webServer.Run(); err != nil {
		log.Fatalf("Ошибка при запуске веб-сервера: %v", err)
	}
}
