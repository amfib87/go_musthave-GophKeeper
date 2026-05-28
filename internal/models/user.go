package models

import (
	"time"

	"github.com/google/uuid"
)

// UserData представляет данные пользователя в системе.
// Содержит метаинформацию и зашифрованные пользовательские данные.
type User struct {
	// ID — уникальный идентификатор записи (UUID).
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

type UserData struct {
	// ID — уникальный идентификатор записи (UUID).
	ID uuid.UUID `json:"id"`
	// UserID — идентификатор пользователя, которому принадлежат данные.
	UserID uuid.UUID `json:"user_id"`
	// DataType — тип данных (например, "password", "card", "text").
	DataType string `json:"data_type"`
	// Data — зашифрованные пользовательские данные (в формате base64).
	Data     []byte `json:"encrypted_data"`
	Metadata string `json:"metadata"`
	// Version — номер версии данных для синхронизации.
	Version int `json:"version"`
	// CreatedAt — timestamp создания записи (Unix time).
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt — timestamp последнего обновления (Unix time).
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt — timestamp удаления (если данные удалены).
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// IsDeleted — флаг удаления записи.
	IsDeleted bool `json:"is_deleted"`
}
