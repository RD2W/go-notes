package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

// Config структура для хранения конфигурации приложения
type Config struct {
	Env      string         `toml:"env" env:"ENV"`
	LogLevel string         `toml:"log_level" env:"LOG_LEVEL"`
	Server   ServerConfig   `toml:"server"`
	Auth     AuthConfig     `toml:"auth"`
	Postgres PostgresConfig `toml:"postgres"`
	Redis    RedisConfig    `toml:"redis"`
	JWT      JWTConfig      `toml:"jwt"`
	Refresh  RefreshConfig  `toml:"refresh"`
	Security SecurityConfig `toml:"security"`
	Shutdown ShutdownConfig `toml:"shutdown"`
}

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Port         string `toml:"port" env:"SERVER_PORT"`
	GRPCPort     string `toml:"grpc_port" env:"GRPC_PORT"`
	ReadTimeout  int    `toml:"read_timeout" env:"SERVER_READ_TIMEOUT"`   // в секундах
	WriteTimeout int    `toml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"` // в секундах
}

// AuthConfig содержит настройки аутентификации
type AuthConfig struct {
	EnableHTTPS bool `toml:"enable_https" env:"AUTH_ENABLE_HTTPS"`
}

// PostgresConfig содержит настройки PostgreSQL
type PostgresConfig struct {
	Host       string `toml:"host" env:"POSTGRES_HOST"`
	Port       int    `toml:"port" env:"POSTGRES_PORT"`
	Name       string `toml:"name" env:"POSTGRES_NAME"`
	User       string `toml:"user" env:"POSTGRES_USER"`
	Password   string `toml:"password" env:"POSTGRES_PASSWORD"`
	SSLMode    string `toml:"ssl_mode" env:"POSTGRES_SSL_MODE"`
	PoolSize   int    `toml:"pool_size" env:"POSTGRES_POOL_SIZE"`
	Parameters string `toml:"parameters" env:"POSTGRES_PARAMETERS"` // дополнительные параметры подключения
}

// RedisConfig содержит настройки Redis
type RedisConfig struct {
	Host     string `toml:"host" env:"REDIS_HOST"`
	Port     int    `toml:"port" env:"REDIS_PORT"`
	Password string `toml:"password" env:"REDIS_PASSWORD"`
	DB       int    `toml:"db" env:"REDIS_DB"`
	PoolSize int    `toml:"pool_size" env:"REDIS_POOL_SIZE"`
	URL      string `toml:"url" env:"REDIS_URL"` // альтернативный способ указания подключения
}

// JWTConfig содержит настройки JWT токенов
type JWTConfig struct {
	SecretKey       string `toml:"secret_key" env:"JWT_SECRET_KEY"`
	Algorithm       string `toml:"algorithm" env:"JWT_ALGORITHM"`
	BcryptCost      int    `toml:"bcrypt_cost" env:"BCRYPT_COST"`             // стоимость хеширования паролей
	AccessTokenTTL  string `toml:"access_token_ttl" env:"ACCESS_TOKEN_TTL"`   // время жизни access токена
	RefreshTokenTTL string `toml:"refresh_token_ttl" env:"REFRESH_TOKEN_TTL"` // время жизни refresh токена
}

// RefreshConfig содержит настройки Refresh токенов
type RefreshConfig struct {
	SecretKey         string `toml:"secret_key" env:"REFRESH_SECRET_KEY"`
	RevocationEnabled bool   `toml:"revocation_enabled" env:"REFRESH_REVOCATION_ENABLED"` // включено ли отслеживание отозванных токенов
}

type SecurityConfig struct {
	PasswordMinLength int    `toml:"password_min_length" env:"PASSWORD_MIN_LENGTH"` // минимальная длина пароля
	MaxLoginAttempts  int    `toml:"max_login_attempts" env:"MAX_LOGIN_ATTEMPTS"`   // максимальное количество попыток входа
	LoginBlockTime    string `toml:"login_block_time" env:"LOGIN_BLOCK_TIME"`       // время блокировки после неудачных попыток
	BcryptCost        int    `toml:"bcrypt_cost" env:"BCRYPT_COST_SEC"`             // стоимость хеширования паролей (дублирует JWT.BcryptCost для удобства)
}

// ShutdownConfig содержит настройки завершения работы
type ShutdownConfig struct {
	Timeout string `toml:"timeout" env:"SHUTDOWN_TIMEOUT"` // таймаут завершения работы
	Wait    string `toml:"wait" env:"SHUTDOWN_WAIT"`       // время ожидания перед завершением
}

// LoadConfig загружает конфигурацию из TOML файла и переменных окружения
func LoadConfig(configPath string) (*Config, error) {
	config := newDefaultConfig()

	// Загружаем конфигурацию из файла
	if err := config.loadFromFile(configPath); err != nil {
		return nil, err
	}

	// Перезаписываем значения из переменных окружения
	config.loadFromEnv()

	return config, nil
}

// loadFromEnv загружает значения из переменных окружения
func (c *Config) loadFromEnv() {
	// General
	if env := os.Getenv("ENV"); env != "" {
		c.Env = env
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	// Server
	if port := os.Getenv("SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		c.Server.GRPCPort = grpcPort
	}
	if readTimeout := os.Getenv("SERVER_READ_TIMEOUT"); readTimeout != "" {
		if val, err := strconv.Atoi(readTimeout); err == nil {
			c.Server.ReadTimeout = val
		}
	}
	if writeTimeout := os.Getenv("SERVER_WRITE_TIMEOUT"); writeTimeout != "" {
		if val, err := strconv.Atoi(writeTimeout); err == nil {
			c.Server.WriteTimeout = val
		}
	}

	// Auth
	if enableHTTPS := os.Getenv("AUTH_ENABLE_HTTPS"); enableHTTPS != "" {
		if val, err := strconv.ParseBool(enableHTTPS); err == nil {
			c.Auth.EnableHTTPS = val
		}
	}

	// Postgres
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		c.Postgres.Host = host
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			c.Postgres.Port = val
		}
	}
	if name := os.Getenv("POSTGRES_NAME"); name != "" {
		c.Postgres.Name = name
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		c.Postgres.User = user
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		c.Postgres.Password = password
	}
	if sslMode := os.Getenv("POSTGRES_SSL_MODE"); sslMode != "" {
		c.Postgres.SSLMode = sslMode
	}
	if poolSize := os.Getenv("POSTGRES_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			c.Postgres.PoolSize = val
		}
	}
	if parameters := os.Getenv("POSTGRES_PARAMETERS"); parameters != "" {
		c.Postgres.Parameters = parameters
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			c.Redis.Port = val
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		c.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if val, err := strconv.Atoi(db); err == nil {
			c.Redis.DB = val
		}
	}
	if poolSize := os.Getenv("REDIS_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			c.Redis.PoolSize = val
		}
	}
	if url := os.Getenv("REDIS_URL"); url != "" {
		c.Redis.URL = url
	}

	// JWT
	if secretKey := os.Getenv("JWT_SECRET_KEY"); secretKey != "" {
		c.JWT.SecretKey = secretKey
	}
	if algorithm := os.Getenv("JWT_ALGORITHM"); algorithm != "" {
		c.JWT.Algorithm = algorithm
	}
	if bcryptCost := os.Getenv("BCRYPT_COST"); bcryptCost != "" {
		if val, err := strconv.Atoi(bcryptCost); err == nil {
			c.JWT.BcryptCost = val
		}
	}
	if accessTokenTTL := os.Getenv("ACCESS_TOKEN_TTL"); accessTokenTTL != "" {
		c.JWT.AccessTokenTTL = accessTokenTTL
	}
	if refreshTokenTTL := os.Getenv("REFRESH_TOKEN_TTL"); refreshTokenTTL != "" {
		c.JWT.RefreshTokenTTL = refreshTokenTTL
	}

	// Refresh
	if secretKey := os.Getenv("REFRESH_SECRET_KEY"); secretKey != "" {
		c.Refresh.SecretKey = secretKey
	}
	if revocationEnabled := os.Getenv("REFRESH_REVOCATION_ENABLED"); revocationEnabled != "" {
		if val, err := strconv.ParseBool(revocationEnabled); err == nil {
			c.Refresh.RevocationEnabled = val
		}
	}
	// Security
	if passwordMinLength := os.Getenv("PASSWORD_MIN_LENGTH"); passwordMinLength != "" {
		if val, err := strconv.Atoi(passwordMinLength); err == nil {
			c.Security.PasswordMinLength = val
		}
	}
	if maxLoginAttempts := os.Getenv("MAX_LOGIN_ATTEMPTS"); maxLoginAttempts != "" {
		if val, err := strconv.Atoi(maxLoginAttempts); err == nil {
			c.Security.MaxLoginAttempts = val
		}
	}
	if loginBlockTime := os.Getenv("LOGIN_BLOCK_TIME"); loginBlockTime != "" {
		c.Security.LoginBlockTime = loginBlockTime
	}
	if bcryptCostSec := os.Getenv("BCRYPT_COST_SEC"); bcryptCostSec != "" {
		if val, err := strconv.Atoi(bcryptCostSec); err == nil {
			c.Security.BcryptCost = val
		}
	}

	// Shutdown
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		c.Shutdown.Timeout = timeout
	}
	if wait := os.Getenv("SHUTDOWN_WAIT"); wait != "" {
		c.Shutdown.Wait = wait
	}
}

// loadFromFile загружает конфигурацию из TOML файла
func (c *Config) loadFromFile(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
		}

		if err := toml.Unmarshal(data, c); err != nil {
			return fmt.Errorf("ошибка парсинга TOML файла конфигурации: %w", err)
		}
	} else if os.IsNotExist(err) {
		log.Printf("Файл конфигурации %s не найден, используются значения по умолчанию", configPath)
	} else {
		return fmt.Errorf("ошибка проверки файла конфигурации: %w", err)
	}

	return nil
}

// newDefaultConfig создает конфигурацию с настройками по умолчанию
func newDefaultConfig() *Config {
	return &Config{
		Env:      "development",
		LogLevel: "info",
		Server: ServerConfig{
			Port:         ":8080",
			GRPCPort:     ":50051",
			ReadTimeout:  15,
			WriteTimeout: 15,
		},
		JWT: JWTConfig{
			SecretKey:       "my_secret_key",
			Algorithm:       "HS256",
			BcryptCost:      10,
			AccessTokenTTL:  "15m",
			RefreshTokenTTL: "168h",
		},
		Refresh: RefreshConfig{
			SecretKey:         "refresh_secret_key",
			RevocationEnabled: true,
		},
		Postgres: PostgresConfig{
			Host:       "localhost",
			Port:       5432,
			Name:       "go_notes",
			User:       "postgres",
			Password:   "notes_password",
			SSLMode:    "disable",
			PoolSize:   10,
			Parameters: "",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
			URL:      "redis://localhost:6379",
		},
		Auth: AuthConfig{
			EnableHTTPS: false,
		},
		Security: SecurityConfig{
			PasswordMinLength: 8,
			MaxLoginAttempts:  5,
			LoginBlockTime:    "30m",
			BcryptCost:        10,
		},
		Shutdown: ShutdownConfig{
			Timeout: "25s",
			Wait:    "3s",
		},
	}
}

// NewDefaultConfigWithValues создает новую конфигурацию с настройками по умолчанию
func NewDefaultConfigWithValues() *Config {
	return newDefaultConfig()
}
