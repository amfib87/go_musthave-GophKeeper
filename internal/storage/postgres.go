// Package storage абстрагирует работу с хранилищем данных (БД).
// Предоставляет унифицированный интерфейс для взаимодействия с разными СУБД.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

type Storage interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserDataByID(ctx context.Context, id uuid.UUID) (*models.UserData, error)
	GetUserAllDataByID(ctx context.Context, id uuid.UUID) ([]*models.UserData, error)
	CreateUserData(ctx context.Context, data *models.UserData) error
	GetUserData(ctx context.Context, userID uuid.UUID) ([]*models.UserData, error)
	UpdateUserData(ctx context.Context, data *models.UserData) error
	Close()
}

type PostgresStorage struct {
	DB *pgxpool.Pool
}

// Initialize устанавливает соединение с базой данных по указанной строке подключения.
// Параметры:
//   - path (string): строка подключения к БД (например, "postgres://user:pass@host/db").
//
// Возвращает:
//   - PostgresStorage: интерфейс хранилища данных;
//   - error: ошибка подключения (если есть).
func Initialize(path string) (*PostgresStorage, error) {
	config, err := pgxpool.ParseConfig(path)
	if err != nil {
		return &PostgresStorage{}, fmt.Errorf("failed pgxpool.ParseConfig %w", err)
	}

	// Настройка параметров пула
	config.MaxConns = 20
	config.MinConns = 5
	config.HealthCheckPeriod = 1 * time.Minute
	config.MaxConnIdleTime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return &PostgresStorage{}, fmt.Errorf("failed pgxpool.NewWithConfig %w", err)
	}

	// Применяем миграции
	if err := runMigrations(pool); err != nil {
		return &PostgresStorage{}, err
	}

	return &PostgresStorage{DB: pool}, nil
}

func runMigrations(pool *pgxpool.Pool) error {
	// Конвертируем pgxpool.Pool в *sql.DB для совместимости с goose
	sqlDB := stdlib.OpenDBFromPool(pool)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed goose.SetDialect: %w", err)
	}

	if err := goose.Up(sqlDB, "../../migrations"); err != nil {
		return fmt.Errorf("failed goose.Up: %w", err)
	}

	return nil
}

func (db *PostgresStorage) Close() {
	db.Close()
}

func (db *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	query := "INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)"

	_, err := db.DB.Exec(ctx, query, user.ID, user.Email, user.Password)
	return err
}

func (s *PostgresStorage) CreateUserData(ctx context.Context, data *models.UserData) error {
	query := `INSERT INTO user_data (id, user_id, data_type, encrypted_data, metadata, version, created_at, updated_at) 
	               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.DB.Exec(ctx, query, data.ID, data.UserID, data.DataType, data.Data, data.Metadata, data.Version, data.CreatedAt, data.UpdatedAt)
	return err
}

func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := "SELECT id, email, password_hash FROM users WHERE email = $1"
	err := s.DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) GetUserData(ctx context.Context, userID uuid.UUID) ([]*models.UserData, error) {
	query := "SELECT id, data_type, encrypted_data, metadata, version, created_at, updated_at FROM user_data WHERE user_id = $1"

	rows, err := s.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*models.UserData
	for rows.Next() {
		var item models.UserData
		err := rows.Scan(&item.ID, &item.DataType, &item.Data, &item.Metadata, &item.Version, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, &item)
	}
	return data, nil
}

// GetUserDataByID получает данные по ID
func (s *PostgresStorage) GetUserDataByID(ctx context.Context, id uuid.UUID) (*models.UserData, error) {

	query := `SELECT id, data_type, encrypted_data, metadata, version, created_at, updated_at, deleted_at, is_deleted 
			    FROM user_data
			WHERE user_id = $1 AND is_deleted = false`

	var data models.UserData
	err := s.DB.QueryRow(ctx, query, id).Scan(
		&data.ID, &data.DataType, &data.Data,
		&data.Metadata, &data.Version, &data.CreatedAt,
		&data.UpdatedAt, &data.DeletedAt, &data.IsDeleted,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *PostgresStorage) GetUserAllDataByID(ctx context.Context, id uuid.UUID) ([]*models.UserData, error) {

	query := `SELECT id, user_id, data_type, encrypted_data, metadata, version, created_at, updated_at, deleted_at
			    FROM user_data
			   WHERE user_id = $1 AND is_deleted = false`

	rows, err := s.DB.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*models.UserData

	for rows.Next() {
		var item models.UserData
		err := rows.Scan(&item.ID, &item.UserID, &item.DataType, &item.Data, &item.Metadata, &item.Version, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil

}

// UpdateUserData обновляет существующие данные
func (s *PostgresStorage) UpdateUserData(ctx context.Context, data *models.UserData) error {
	query := `UPDATE user_data
				 SET encrypted_data = $2, metadata = $3, version = $4, updated_at = $5, deleted_at = $6, is_deleted = $7
		       WHERE user_id = $8 AND data_type = $1`

	_, err := s.DB.Exec(ctx, query, data.DataType, data.Data, data.Metadata, data.Version, data.UpdatedAt, data.DeletedAt, data.IsDeleted, data.UserID)
	return err
}
