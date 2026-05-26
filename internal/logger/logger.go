// Package logger предоставляет функционал для структурированного логирования
// с использованием библиотеки Zap.
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Tlog struct {
	Log *zap.Logger
}

// Initialize создаёт и настраивает экземпляр логгера Zap.
// Возвращает:
//   - *Logger: инициализированный логгер;
//   - error: ошибка инициализации (если есть).
func Initialize() (Tlog, error) {
	log, err := zap.NewDevelopment()
	if err != nil {
		return Tlog{}, fmt.Errorf("failed zap.NewDevelopment %w", err)
	}

	return Tlog{Log: log}, nil
}
