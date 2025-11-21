package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Создаем временный файл конфигурации для теста
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.toml")

	// Создаем тестовый TOML файл
	testConfigContent := `env = "test"
log_level = "debug"

[server]
port = ":9090"
grpc_port = ":50052"
read_timeout = 20
write_timeout = 25

[auth]
enable_https = true

[postgres]
host = "test_host"
port = 5433
name = "test_db"
user = "test_user"
password = "test_password"
ssl_mode = "require"
pool_size = 5
parameters = "test_param=value"

[redis]
host = "test_redis_host"
port = 6380
password = "test_redis_password"
db = 1
pool_size = 5
url = "redis://test_redis_host:6380"

[jwt]
secret_key = "test_secret_key"
algorithm = "HS512"
bcrypt_cost = 12
access_token_ttl = "30m"
refresh_token_ttl = "72h"

[refresh]
secret_key = "test_refresh_secret_key"
revocation_enabled = false

[security]
password_min_length = 10
max_login_attempts = 3
login_block_time = "15m"
bcrypt_cost = 12

[shutdown]
timeout = "30s"
wait = "5s"
`

	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать временный файл конфигурации: %v", err)
	}

	t.Run("Load config from file", func(t *testing.T) {
		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig вернула ошибку: %v", err)
		}

		// Проверяем, что значения загружены из файла
		expected := &Config{
			Env:      "test",
			LogLevel: "debug",
			Server: ServerConfig{
				Port:         ":9090",
				GRPCPort:     ":50052",
				ReadTimeout:  20,
				WriteTimeout: 25,
			},
			Auth: AuthConfig{
				EnableHTTPS: true,
			},
			Postgres: PostgresConfig{
				Host:       "test_host",
				Port:       5433,
				Name:       "test_db",
				User:       "test_user",
				Password:   "test_password",
				SSLMode:    "require",
				PoolSize:   5,
				Parameters: "test_param=value",
			},
			Redis: RedisConfig{
				Host:     "test_redis_host",
				Port:     6380,
				Password: "test_redis_password",
				DB:       1,
				PoolSize: 5,
				URL:      "redis://test_redis_host:6380",
			},
			JWT: JWTConfig{
				SecretKey:       "test_secret_key",
				Algorithm:       "HS512",
				BcryptCost:      12,
				AccessTokenTTL:  "30m",
				RefreshTokenTTL: "72h",
			},
			Refresh: RefreshConfig{
				SecretKey:         "test_refresh_secret_key",
				RevocationEnabled: false,
			},
			Security: SecurityConfig{
				PasswordMinLength: 10,
				MaxLoginAttempts:  3,
				LoginBlockTime:    "15m",
				BcryptCost:        12,
			},
			Shutdown: ShutdownConfig{
				Timeout: "30s",
				Wait:    "5s",
			},
		}

		if !reflect.DeepEqual(config, expected) {
			t.Errorf("Config не соответствует ожидаемому значению.\nПолучено: %+v\nОжидается: %+v", config, expected)
		}
	})

	t.Run("Load config with environment variable overrides", func(t *testing.T) {
		// Устанавливаем переменные окружения для переопределения
		if err := os.Setenv("ENV", "production"); err != nil {
			t.Fatalf("Не удалось установить переменную окружения ENV: %v", err)
		}
		if err := os.Setenv("SERVER_PORT", ":8081"); err != nil {
			t.Fatalf("Не удалось установить переменную окружения SERVER_PORT: %v", err)
		}
		if err := os.Setenv("POSTGRES_HOST", "prod_host"); err != nil {
			t.Fatalf("Не удалось установить переменную окружения POSTGRES_HOST: %v", err)
		}
		defer func() {
			// Очищаем переменные окружения после теста
			if err := os.Unsetenv("ENV"); err != nil {
				t.Errorf("Не удалось очистить переменную окружения ENV: %v", err)
			}
			if err := os.Unsetenv("SERVER_PORT"); err != nil {
				t.Errorf("Не удалось очистить переменную окружения SERVER_PORT: %v", err)
			}
			if err := os.Unsetenv("POSTGRES_HOST"); err != nil {
				t.Errorf("Не удалось очистить переменную окружения POSTGRES_HOST: %v", err)
			}
		}()

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig вернула ошибку: %v", err)
		}

		// Проверяем, что значения были переопределены из переменных окружения
		if config.Env != "production" {
			t.Errorf("Env должно быть 'production', получено: %s", config.Env)
		}
		if config.Server.Port != ":8081" {
			t.Errorf("Server.Port должно быть ':8081', получено: %s", config.Server.Port)
		}
		if config.Postgres.Host != "prod_host" {
			t.Errorf("Postgres.Host должно быть 'prod_host', получено: %s", config.Postgres.Host)
		}

		// Проверяем, что остальные значения остались из файла
		if config.LogLevel != "debug" {
			t.Errorf("LogLevel должно быть 'debug', получено: %s", config.LogLevel)
		}
	})
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("nonexistent_config.toml")
	if err != nil {
		t.Fatalf("LoadConfig должна использовать значения по умолчанию при отсутствии файла, а не возвращать ошибку: %v", err)
	}

	// Проверяем, что используется конфигурация по умолчанию
	defaultConfig := newDefaultConfig()
	if !reflect.DeepEqual(config, defaultConfig) {
		t.Errorf("Config не соответствует конфигурации по умолчанию.\nПолучено: %+v\nОжидается: %+v", config, defaultConfig)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Устанавливаем переменные окружения
	if err := os.Setenv("ENV", "test_env"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения ENV: %v", err)
	}
	if err := os.Setenv("SERVER_PORT", ":9999"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения SERVER_PORT: %v", err)
	}
	if err := os.Setenv("GRPC_PORT", ":60061"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения GRPC_PORT: %v", err)
	}
	if err := os.Setenv("SERVER_READ_TIMEOUT", "30"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения SERVER_READ_TIMEOUT: %v", err)
	}
	if err := os.Setenv("SERVER_WRITE_TIMEOUT", "40"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения SERVER_WRITE_TIMEOUT: %v", err)
	}
	if err := os.Setenv("AUTH_ENABLE_HTTPS", "true"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения AUTH_ENABLE_HTTPS: %v", err)
	}
	if err := os.Setenv("POSTGRES_HOST", "env_host"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_HOST: %v", err)
	}
	if err := os.Setenv("POSTGRES_PORT", "5435"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_PORT: %v", err)
	}
	if err := os.Setenv("POSTGRES_NAME", "env_db"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_NAME: %v", err)
	}
	if err := os.Setenv("POSTGRES_USER", "env_user"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_USER: %v", err)
	}
	if err := os.Setenv("POSTGRES_PASSWORD", "env_password"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_PASSWORD: %v", err)
	}
	if err := os.Setenv("POSTGRES_SSL_MODE", "require"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_SSL_MODE: %v", err)
	}
	if err := os.Setenv("POSTGRES_POOL_SIZE", "15"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_POOL_SIZE: %v", err)
	}
	if err := os.Setenv("POSTGRES_PARAMETERS", "env_param=value"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения POSTGRES_PARAMETERS: %v", err)
	}
	if err := os.Setenv("REDIS_HOST", "env_redis_host"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_HOST: %v", err)
	}
	if err := os.Setenv("REDIS_PORT", "6381"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_PORT: %v", err)
	}
	if err := os.Setenv("REDIS_PASSWORD", "env_redis_password"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_PASSWORD: %v", err)
	}
	if err := os.Setenv("REDIS_DB", "2"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_DB: %v", err)
	}
	if err := os.Setenv("REDIS_POOL_SIZE", "8"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_POOL_SIZE: %v", err)
	}
	if err := os.Setenv("REDIS_URL", "redis://env_redis_host:6381"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REDIS_URL: %v", err)
	}
	if err := os.Setenv("JWT_SECRET_KEY", "env_secret_key"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения JWT_SECRET_KEY: %v", err)
	}
	if err := os.Setenv("JWT_ALGORITHM", "RS256"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения JWT_ALGORITHM: %v", err)
	}
	if err := os.Setenv("BCRYPT_COST", "14"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения BCRYPT_COST: %v", err)
	}
	if err := os.Setenv("ACCESS_TOKEN_TTL", "20m"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения ACCESS_TOKEN_TTL: %v", err)
	}
	if err := os.Setenv("REFRESH_TOKEN_TTL", "200h"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REFRESH_TOKEN_TTL: %v", err)
	}
	if err := os.Setenv("REFRESH_SECRET_KEY", "env_refresh_secret_key"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REFRESH_SECRET_KEY: %v", err)
	}
	if err := os.Setenv("REFRESH_REVOCATION_ENABLED", "false"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения REFRESH_REVOCATION_ENABLED: %v", err)
	}
	if err := os.Setenv("PASSWORD_MIN_LENGTH", "12"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения PASSWORD_MIN_LENGTH: %v", err)
	}
	if err := os.Setenv("MAX_LOGIN_ATTEMPTS", "2"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения MAX_LOGIN_ATTEMPTS: %v", err)
	}
	if err := os.Setenv("LOGIN_BLOCK_TIME", "45m"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения LOGIN_BLOCK_TIME: %v", err)
	}
	if err := os.Setenv("BCRYPT_COST_SEC", "14"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения BCRYPT_COST_SEC: %v", err)
	}
	if err := os.Setenv("SHUTDOWN_TIMEOUT", "40s"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения SHUTDOWN_TIMEOUT: %v", err)
	}
	if err := os.Setenv("SHUTDOWN_WAIT", "7s"); err != nil {
		t.Fatalf("Не удалось установить переменную окружения SHUTDOWN_WAIT: %v", err)
	}
	defer func() {
		// Очищаем переменные окружения после теста
		if err := os.Unsetenv("ENV"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения ENV: %v", err)
		}
		if err := os.Unsetenv("SERVER_PORT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения SERVER_PORT: %v", err)
		}
		if err := os.Unsetenv("GRPC_PORT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения GRPC_PORT: %v", err)
		}
		if err := os.Unsetenv("SERVER_READ_TIMEOUT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения SERVER_READ_TIMEOUT: %v", err)
		}
		if err := os.Unsetenv("SERVER_WRITE_TIMEOUT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения SERVER_WRITE_TIMEOUT: %v", err)
		}
		if err := os.Unsetenv("AUTH_ENABLE_HTTPS"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения AUTH_ENABLE_HTTPS: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_HOST"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_HOST: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_PORT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_PORT: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_NAME"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_NAME: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_USER"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_USER: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_PASSWORD"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_PASSWORD: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_SSL_MODE"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_SSL_MODE: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_POOL_SIZE"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_POOL_SIZE: %v", err)
		}
		if err := os.Unsetenv("POSTGRES_PARAMETERS"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения POSTGRES_PARAMETERS: %v", err)
		}
		if err := os.Unsetenv("REDIS_HOST"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_HOST: %v", err)
		}
		if err := os.Unsetenv("REDIS_PORT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_PORT: %v", err)
		}
		if err := os.Unsetenv("REDIS_PASSWORD"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_PASSWORD: %v", err)
		}
		if err := os.Unsetenv("REDIS_DB"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_DB: %v", err)
		}
		if err := os.Unsetenv("REDIS_POOL_SIZE"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_POOL_SIZE: %v", err)
		}
		if err := os.Unsetenv("REDIS_URL"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REDIS_URL: %v", err)
		}
		if err := os.Unsetenv("JWT_SECRET_KEY"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения JWT_SECRET_KEY: %v", err)
		}
		if err := os.Unsetenv("JWT_ALGORITHM"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения JWT_ALGORITHM: %v", err)
		}
		if err := os.Unsetenv("BCRYPT_COST"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения BCRYPT_COST: %v", err)
		}
		if err := os.Unsetenv("ACCESS_TOKEN_TTL"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения ACCESS_TOKEN_TTL: %v", err)
		}
		if err := os.Unsetenv("REFRESH_TOKEN_TTL"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REFRESH_TOKEN_TTL: %v", err)
		}
		if err := os.Unsetenv("REFRESH_SECRET_KEY"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REFRESH_SECRET_KEY: %v", err)
		}
		if err := os.Unsetenv("REFRESH_REVOCATION_ENABLED"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения REFRESH_REVOCATION_ENABLED: %v", err)
		}
		if err := os.Unsetenv("PASSWORD_MIN_LENGTH"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения PASSWORD_MIN_LENGTH: %v", err)
		}
		if err := os.Unsetenv("MAX_LOGIN_ATTEMPTS"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения MAX_LOGIN_ATTEMPTS: %v", err)
		}
		if err := os.Unsetenv("LOGIN_BLOCK_TIME"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения LOGIN_BLOCK_TIME: %v", err)
		}
		if err := os.Unsetenv("BCRYPT_COST_SEC"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения BCRYPT_COST_SEC: %v", err)
		}
		if err := os.Unsetenv("SHUTDOWN_TIMEOUT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения SHUTDOWN_TIMEOUT: %v", err)
		}
		if err := os.Unsetenv("SHUTDOWN_WAIT"); err != nil {
			t.Errorf("Не удалось очистить переменную окружения SHUTDOWN_WAIT: %v", err)
		}
	}()

	config := newDefaultConfig()
	config.loadFromEnv()

	// Проверяем, что значения были установлены из переменных окружения
	if config.Env != "test_env" {
		t.Errorf("Env должно быть 'test_env', получено: %s", config.Env)
	}
	if config.Server.Port != ":9999" {
		t.Errorf("Server.Port должно быть ':9999', получено: %s", config.Server.Port)
	}
	if config.Server.GRPCPort != ":60061" {
		t.Errorf("Server.GRPCPort должно быть ':60061', получено: %s", config.Server.GRPCPort)
	}
	if config.Server.ReadTimeout != 30 {
		t.Errorf("Server.ReadTimeout должно быть 30, получено: %d", config.Server.ReadTimeout)
	}
	if config.Server.WriteTimeout != 40 {
		t.Errorf("Server.WriteTimeout должно быть 40, получено: %d", config.Server.WriteTimeout)
	}
	if !config.Auth.EnableHTTPS {
		t.Errorf("Auth.EnableHTTPS должно быть true, получено: %t", config.Auth.EnableHTTPS)
	}
	if config.Postgres.Host != "env_host" {
		t.Errorf("Postgres.Host должно быть 'env_host', получено: %s", config.Postgres.Host)
	}
	if config.Postgres.Port != 5435 {
		t.Errorf("Postgres.Port должно быть 5435, получено: %d", config.Postgres.Port)
	}
	if config.Postgres.Name != "env_db" {
		t.Errorf("Postgres.Name должно быть 'env_db', получено: %s", config.Postgres.Name)
	}
	if config.Postgres.User != "env_user" {
		t.Errorf("Postgres.User должно быть 'env_user', получено: %s", config.Postgres.User)
	}
	if config.Postgres.Password != "env_password" {
		t.Errorf("Postgres.Password должно быть 'env_password', получено: %s", config.Postgres.Password)
	}
	if config.Postgres.SSLMode != "require" {
		t.Errorf("Postgres.SSLMode должно быть 'require', получено: %s", config.Postgres.SSLMode)
	}
	if config.Postgres.PoolSize != 15 {
		t.Errorf("Postgres.PoolSize должно быть 15, получено: %d", config.Postgres.PoolSize)
	}
	if config.Postgres.Parameters != "env_param=value" {
		t.Errorf("Postgres.Parameters должно быть 'env_param=value', получено: %s", config.Postgres.Parameters)
	}
	if config.Redis.Host != "env_redis_host" {
		t.Errorf("Redis.Host должно быть 'env_redis_host', получено: %s", config.Redis.Host)
	}
	if config.Redis.Port != 6381 {
		t.Errorf("Redis.Port должно быть 6381, получено: %d", config.Redis.Port)
	}
	if config.Redis.Password != "env_redis_password" {
		t.Errorf("Redis.Password должно быть 'env_redis_password', получено: %s", config.Redis.Password)
	}
	if config.Redis.DB != 2 {
		t.Errorf("Redis.DB должно быть 2, получено: %d", config.Redis.DB)
	}
	if config.Redis.PoolSize != 8 {
		t.Errorf("Redis.PoolSize должно быть 8, получено: %d", config.Redis.PoolSize)
	}
	if config.Redis.URL != "redis://env_redis_host:6381" {
		t.Errorf("Redis.URL должно быть 'redis://env_redis_host:6381', получено: %s", config.Redis.URL)
	}
	if config.JWT.SecretKey != "env_secret_key" {
		t.Errorf("JWT.SecretKey должно быть 'env_secret_key', получено: %s", config.JWT.SecretKey)
	}
	if config.JWT.Algorithm != "RS256" {
		t.Errorf("JWT.Algorithm должно быть 'RS256', получено: %s", config.JWT.Algorithm)
	}
	if config.JWT.BcryptCost != 14 {
		t.Errorf("JWT.BcryptCost должно быть 14, получено: %d", config.JWT.BcryptCost)
	}
	if config.JWT.AccessTokenTTL != "20m" {
		t.Errorf("JWT.AccessTokenTTL должно быть '20m', получено: %s", config.JWT.AccessTokenTTL)
	}
	if config.JWT.RefreshTokenTTL != "200h" {
		t.Errorf("JWT.RefreshTokenTTL должно быть '200h', получено: %s", config.JWT.RefreshTokenTTL)
	}
	if config.Refresh.SecretKey != "env_refresh_secret_key" {
		t.Errorf("Refresh.SecretKey должно быть 'env_refresh_secret_key', получено: %s", config.Refresh.SecretKey)
	}
	if config.Refresh.RevocationEnabled {
		t.Errorf("Refresh.RevocationEnabled должно быть false, получено: %t", config.Refresh.RevocationEnabled)
	}
	if config.Security.PasswordMinLength != 12 {
		t.Errorf("Security.PasswordMinLength должно быть 12, получено: %d", config.Security.PasswordMinLength)
	}
	if config.Security.MaxLoginAttempts != 2 {
		t.Errorf("Security.MaxLoginAttempts должно быть 2, получено: %d", config.Security.MaxLoginAttempts)
	}
	if config.Security.LoginBlockTime != "45m" {
		t.Errorf("Security.LoginBlockTime должно быть '45m', получено: %s", config.Security.LoginBlockTime)
	}
	if config.Security.BcryptCost != 14 {
		t.Errorf("Security.BcryptCost должно быть 14, получено: %d", config.Security.BcryptCost)
	}
	if config.Shutdown.Timeout != "40s" {
		t.Errorf("Shutdown.Timeout должно быть '40s', получено: %s", config.Shutdown.Timeout)
	}
	if config.Shutdown.Wait != "7s" {
		t.Errorf("Shutdown.Wait должно быть '7s', получено: %s", config.Shutdown.Wait)
	}
}

func TestLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "load_from_file_test.toml")

	testConfigContent := `env = "file_test"
log_level = "warn"

[server]
port = ":7777"
grpc_port = ":55555"
read_timeout = 25
write_timeout = 30
`

	err := os.WriteFile(configPath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать временный файл конфигурации: %v", err)
	}

	config := newDefaultConfig()
	err = config.loadFromFile(configPath)
	if err != nil {
		t.Fatalf("loadFromFile вернула ошибку: %v", err)
	}

	// Проверяем, что значения были загружены из файла
	if config.Env != "file_test" {
		t.Errorf("Env должно быть 'file_test', получено: %s", config.Env)
	}
	if config.LogLevel != "warn" {
		t.Errorf("LogLevel должно быть 'warn', получено: %s", config.LogLevel)
	}
	if config.Server.Port != ":7777" {
		t.Errorf("Server.Port должно быть ':7777', получено: %s", config.Server.Port)
	}
	if config.Server.GRPCPort != ":55555" {
		t.Errorf("Server.GRPCPort должно быть ':55555', получено: %s", config.Server.GRPCPort)
	}
	if config.Server.ReadTimeout != 25 {
		t.Errorf("Server.ReadTimeout должно быть 25, получено: %d", config.Server.ReadTimeout)
	}
	if config.Server.WriteTimeout != 30 {
		t.Errorf("Server.WriteTimeout должно быть 30, получено: %d", config.Server.WriteTimeout)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	config := newDefaultConfig()
	originalConfig := *config // Сохраняем копию до вызова

	err := config.loadFromFile("nonexistent_file.toml")
	if err != nil {
		t.Fatalf("loadFromFile должна возвращать nil при отсутствии файла, а не ошибку: %v", err)
	}

	// Проверяем, что конфигурация не изменилась
	if !reflect.DeepEqual(config, &originalConfig) {
		t.Errorf("Config изменилась при отсутствии файла. Было: %+v, стало: %+v", &originalConfig, config)
	}
}

func TestNewDefaultConfig(t *testing.T) {
	config := newDefaultConfig()

	// Проверяем, что все поля имеют ожидаемые значения по умолчанию
	expected := &Config{
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

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Конфигурация по умолчанию не соответствует ожидаемой.\nПолучено: %+v\nОжидается: %+v", config, expected)
	}
}

func TestNewDefaultConfigWithValues(t *testing.T) {
	config1 := NewDefaultConfigWithValues()
	config2 := newDefaultConfig()

	// Проверяем, что обе функции возвращают эквивалентные значения
	if !reflect.DeepEqual(config1, config2) {
		t.Errorf("NewDefaultConfigWithValues и newDefaultConfig возвращают разные значения.\nNewDefaultConfigWithValues: %+v\nnewDefaultConfig: %+v", config1, config2)
	}

	// Проверяем, что возвращаемое значение не является nil
	if config1 == nil {
		t.Error("NewDefaultConfigWithValues должна возвращать ненулевое значение")
	}
}
