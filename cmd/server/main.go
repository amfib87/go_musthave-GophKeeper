// Package main — точка входа в приложение GophKeeper.
//
// GophKeeper — серверная часть системы для безопасного хранения пользовательских данных:
// паролей, заметок, банковских карт и другой конфиденциальной информации.
//
// Основные компоненты:
//   - HTTP‑сервер с REST API;
//   - аутентификация и авторизация через JWT;
//   - хранение данных в реляционной БД;
//   - структурированное логирование через Zap.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/amfib87/go_musthave-GophKeeper/internal/handlers"
	"github.com/amfib87/go_musthave-GophKeeper/internal/logger"
	"github.com/amfib87/go_musthave-GophKeeper/internal/router"
	"github.com/amfib87/go_musthave-GophKeeper/internal/storage"
	"go.uber.org/zap"
)

// main — точка входа в приложение.
// Вызывает функцию run() для инициализации и запуска сервера.
// При возникновении ошибки выводит сообщение в лог и завершает программу с кодом ошибки.
func main() {
	if err := run(); err != nil {
		log.Fatal("runtime error:", err)
	}
}

// run выполняет последовательную инициализацию компонентов приложения и запускает HTTP‑сервер.
//
// Порядок инициализации:
// 1. Логгер (logger.Initialize).
// 2. Конфигурация (config.NewConfig, cfg.ParseFlag).
// 3. Хранилище данных (storage.Initialize).
// 4. HTTP‑обработчик (handlers.Newhandler).
// 5. Роутер (router.Initialize).
// 6. Запуск сервера (http.ListenAndServe).
//
// Возвращает:
// - nil при успешном запуске сервера;
// - ошибку, если сбой произошёл на любом этапе инициализации.
func run() error {

	// Создаем логгер
	logger, err := logger.Initialize()
	if err != nil {
		return err
	}
	logger.Log.Debug("logger was successfulle created")

	// Парсим флаги
	cfg := config.NewConfig()
	cfg.ParseFlag()
	logger.Log.Debug("flags were parsed", zap.Any("cfg", cfg))

	// Инициализируем БД
	db, err := storage.Initialize(cfg.DBURI)
	if err != nil {
		return err
	}

	// Инициализируем хэндлер
	handler := handlers.Newhandler(db, cfg, logger)
	defer handler.Close()

	// Инициализируем роутер
	router, err := router.Initialize(handler)
	if err != nil {
		logger.Log.Error("router.Initialize:", zap.Error(err))
		return fmt.Errorf("router.Initialize: %w", err)
	}

	logger.Log.Info("running server", zap.String("cfg.RunAddress)", cfg.ServRunAddr))
	if err := http.ListenAndServe(cfg.ServRunAddr, router); err != nil {
		logger.Log.Error("failed http.ListenAndServe: %v", zap.Error(err))
		return err
	}

	return nil
}
