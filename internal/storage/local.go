package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/google/uuid"
)

type LocalStorage struct {
	dataDir string
}

// NewLocalStorage создаёт новый экземпляр локального хранилища.
// Параметр:
//
//	dataDir — путь к директории для хранения данных.
//
// Возвращает:
//
//	*LocalStorage — инициализированное хранилище.
func NewLocalStorage(dataDir string) *LocalStorage {
	return &LocalStorage{
		dataDir: dataDir,
	}
}

// SaveLocalData сохраняет данные в локальное хранилище.
// Параметр:
//
//	data — данные для сохранения.
//
// Возвращает:
//
//	error — ошибка при сохранении (nil при успехе).
func (ls *LocalStorage) SaveLocalData(data []*models.UserData) error {
	// Создаём директорию для хранения данных, если её нет
	dataDir := filepath.Join(ls.dataDir, "data")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Обрабатываем каждую запись данных
	for _, item := range data {
		// Формируем путь к файлу: ./data/{id}.json
		filePath := filepath.Join(dataDir, fmt.Sprintf("%s.json", item.ID.String()))

		// Подготавливаем данные для сохранения
		saveData := struct {
			ID        uuid.UUID  `json:"id"`
			UserID    uuid.UUID  `json:"user_id"`
			DataType  string     `json:"data_type"`
			Data      string     `json:"data"` // кодируем в base64 для безопасного хранения
			Metadata  string     `json:"metadata,omitempty"`
			Version   int        `json:"version"`
			CreatedAt time.Time  `json:"created_at"`
			UpdatedAt time.Time  `json:"updated_at"`
			DeletedAt *time.Time `json:"deleted_at,omitempty"`
			IsDeleted bool       `json:"is_deleted"`
		}{
			ID:        item.ID,
			UserID:    item.UserID,
			DataType:  item.DataType,
			Data:      base64.StdEncoding.EncodeToString(item.Data),
			Metadata:  item.Metadata,
			Version:   item.Version,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
			IsDeleted: item.IsDeleted,
		}

		// Кодируем в JSON
		jsonData, err := json.MarshalIndent(saveData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal data for ID %s: %w", item.ID, err)
		}

		// Сохраняем в файл с правами 0600 (только для владельца)
		err = os.WriteFile(filePath, jsonData, 0600)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return nil
}

// GetLocalData загружает все данные из локального хранилища.
// Возвращает:
//
//	[]*models.UserData — загруженные данные.
//	error — ошибка при чтении (nil при успехе).
func (ls *LocalStorage) GetLocalData() ([]*models.UserData, error) {
	var result []*models.UserData
	dataDir := filepath.Join(ls.dataDir, "data")

	// Проверяем существование директории
	_, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		return result, nil // директория не существует — возвращаем пустой срез
	} else if err != nil {
		return nil, fmt.Errorf("failed to access data directory: %w", err)
	}

	// Читаем все файлы в директории
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(dataDir, file.Name())

			// Читаем файл
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			// Парсим JSON
			var saveData struct {
				ID        uuid.UUID  `json:"id"`
				DataType  string     `json:"data_type"`
				Data      string     `json:"data"`
				Metadata  string     `json:"metadata,omitempty"`
				Version   int        `json:"version"`
				CreatedAt time.Time  `json:"created_at"`
				UpdatedAt time.Time  `json:"updated_at"`
				DeletedAt *time.Time `json:"deleted_at,omitempty"`
				IsDeleted bool       `json:"is_deleted"`
			}

			err = json.Unmarshal(fileData, &saveData)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal file %s: %w", filePath, err)
			}

			// Декодируем base64 обратно в []byte
			dataBytes, err := base64.StdEncoding.DecodeString(saveData.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to decode data from file %s: %w", filePath, err)
			}

			// Создаём объект UserData
			userData := &models.UserData{
				ID:        saveData.ID,
				DataType:  saveData.DataType,
				Data:      dataBytes,
				Metadata:  saveData.Metadata,
				Version:   saveData.Version,
				CreatedAt: saveData.CreatedAt,
				UpdatedAt: saveData.UpdatedAt,
				DeletedAt: saveData.DeletedAt,
				IsDeleted: saveData.IsDeleted,
			}

			result = append(result, userData)
		}
	}

	return result, nil
}
