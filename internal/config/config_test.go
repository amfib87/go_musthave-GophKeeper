package config_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
)

func TestNewConfig(t *testing.T) {
	cfg := config.NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}

	if cfg.ServRunAddr != "" {
		t.Errorf("expected ServRunAddr to be empty, got %q", cfg.ServRunAddr)
	}
	if cfg.DBURI != "" {
		t.Errorf("expected DBURI to be empty, got %q", cfg.DBURI)
	}
}

func TestConfig_ParseFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		envVars       map[string]string
		expectedAddr  string
		expectedDBURI string
	}{
		{
			name:          "flags only",
			args:          []string{"-a", "localhost:8080", "-b", "postgres://localhost/db"},
			envVars:       map[string]string{},
			expectedAddr:  "localhost:8080",
			expectedDBURI: "postgres://localhost/db",
		},
		{
			name: "env vars only",
			args: []string{},
			envVars: map[string]string{
				"RUN_ADDRESS":  "localhost:9000",
				"DATABASE_URI": "mysql://localhost/db",
			},
			expectedAddr:  "localhost:9000",
			expectedDBURI: "mysql://localhost/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очистка флагов перед тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Сохраняем оригинальные значения переменных окружения
			originalEnv := make(map[string]string)
			for k := range tt.envVars {
				originalEnv[k] = os.Getenv(k)
			}

			// Устанавливаем переменные окружения для теста
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Восстанавливаем оригинальные значения после теста
			defer func() {
				for k, v := range originalEnv {
					os.Setenv(k, v)
				}
			}()

			cfg := config.NewConfig()
			oldArgs := os.Args
			os.Args = append([]string{"cmd"}, tt.args...)
			defer func() { os.Args = oldArgs }()

			// Вызов ParseFlag после настройки окружения
			cfg.ParseFlag()

			if cfg.ServRunAddr != tt.expectedAddr {
				t.Errorf("ServRunAddr: expected %q, got %q", tt.expectedAddr, cfg.ServRunAddr)
			}
			if cfg.DBURI != tt.expectedDBURI {
				t.Errorf("DBURI: expected %q, got %q", tt.expectedDBURI, cfg.DBURI)
			}
		})
	}
}

func TestSaveAuthToken(t *testing.T) {
	tempDir := t.TempDir()

	// Мокаем локальную функцию
	originalFunc := config.UserHomeDirFunc
	config.UserHomeDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { config.UserHomeDirFunc = originalFunc }()

	token := "test-auth-token-123"

	err := config.SaveAuthToken(token)
	if err != nil {
		t.Fatalf("SaveAuthToken() returned unexpected error: %v", err)
	}

	// Проверяем, что файл создан
	configPath := filepath.Join(tempDir, ".gophkeeper")
	tokenPath := filepath.Join(configPath, "auth_token")

	_, err = os.Stat(tokenPath)
	if os.IsNotExist(err) {
		t.Error("token file was not created")
	} else if err != nil {
		t.Errorf("unexpected error checking token file: %v", err)
	}

	// Проверяем содержимое файла
	content, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("failed to read token file: %v", err)
	}
	if string(content) != token {
		t.Errorf("token file content: expected %q, got %q", token, string(content))
	}
}

func TestGetDataDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Мокаем локальную функцию
	originalFunc := config.UserHomeDirFunc
	config.UserHomeDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { config.UserHomeDirFunc = originalFunc }()

	expectedPath := filepath.Join(tempDir, ".gophkeeper", "data")
	actualPath := config.GetDataDirectory()

	if actualPath != expectedPath {
		t.Errorf("GetDataDirectory(): expected %q, got %q", expectedPath, actualPath)
	}
}
