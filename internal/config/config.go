// Package config отвечает за работу с конфигурацией приложения:
// парсинг флагов командной строки, переменных окружения и значений по умолчанию.
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config представляет конфигурацию приложения GophKeeper.
type Config struct {
	// ServRunAddr — URL сервера API.
	ServRunAddr    string
	DBURI          string
	MasterPassword string
}

// NewConfig создаёт новый экземпляр конфигурации с значениями по умолчанию.
// Возвращает указатель на структуру Config.
func NewConfig() *Config {
	return &Config{}
}

var UserHomeDirFunc = os.UserHomeDir

func GetUserHomeDir() (string, error) {
	return UserHomeDirFunc()
}

// ParseFlag парсит флаги командной строки и обновляет поля структуры Config.
// Поддерживаемые флаги:
//   - -run-address: адрес и порт для запуска сервера (по умолчанию: ":8080");
//   - -database-uri: строка подключения к БД.
func (cfg *Config) ParseFlag() {

	flag.StringVar(&cfg.ServRunAddr, "a", "localhost:8080", "address for start server")
	flag.StringVar(&cfg.DBURI, "b", "", "address fot connect to DB")
	flag.StringVar(&cfg.MasterPassword, "p", "MasterPassword", "master password")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envServRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.ServRunAddr = envServRunAddr
	}
	if envDBURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DBURI = envDBURI
	}
	if envPassword, ok := os.LookupEnv("MASTER_PASSWORD"); ok {
		cfg.MasterPassword = envPassword
	}
}

// SaveAuthToken сохраняет токен аутентификации в конфигурационный файл
func SaveAuthToken(token string) error {
	homeDir, err := UserHomeDirFunc()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".gophkeeper")
	err = os.MkdirAll(configPath, 0700) // права для директории
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(configPath, "auth_token")
	// Явно указываем права 0600
	err = os.WriteFile(tokenPath, []byte(token), 0600)
	if err != nil {
		return err
	}
	return nil
}

const (
	appName    = ".gophkeeper"
	dataSubdir = "data"
)

// GetDataDirectory возвращает путь к директории данных приложения
func GetDataDirectory() string {
	home, err := UserHomeDirFunc()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, appName, dataSubdir)
}

func LoadAuthToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tokenPath := filepath.Join(home, ".gophkeeper", "auth_token")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("token file not found: %v", err)
	}
	return string(data), nil
}
