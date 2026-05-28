package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/amfib87/go_musthave-GophKeeper/internal/logger"
	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/amfib87/go_musthave-GophKeeper/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockStorage struct {
	storage.Storage
	createUserCalled     bool
	getUserByEmailCalled bool
	userData             []*models.UserData
	user                 *models.User
}

func (m *mockStorage) CreateUser(ctx context.Context, user *models.User) error {
	m.createUserCalled = true
	m.user = user
	return nil
}

func (m *mockStorage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.getUserByEmailCalled = true
	return m.user, nil
}

func (m *mockStorage) GetUserData(ctx context.Context, userID uuid.UUID) ([]*models.UserData, error) {
	return m.userData, nil
}

func (m *mockStorage) CreateUserData(ctx context.Context, data *models.UserData) error {
	m.userData = append(m.userData, data)
	return nil
}

func (m *mockStorage) UpdateUserData(ctx context.Context, data *models.UserData) error {
	if data == nil || data.ID == uuid.Nil {
		return errors.New("invalid data ID")
	}
	m.userData = append(m.userData, data)
	return nil
}

func (m *mockStorage) GetUserDataByID(ctx context.Context, id uuid.UUID) (*models.UserData, error) {

	for _, item := range m.userData {
		if item.ID == id {
			return item, nil
		}
	}
	return &models.UserData{}, nil
}

func (m *mockStorage) GetUserAllDataByID(ctx context.Context, id uuid.UUID) ([]*models.UserData, error) {
	var result []*models.UserData
	item := &models.UserData{
		UserID: id}

	result = append(result, item)

	return result, nil
}

func TestHandler_Register(t *testing.T) {
	// Настраиваем мок‑хранилище
	mockSt := &mockStorage{}

	// Создаём логгер
	logger, err := logger.Initialize()
	if err != nil {
		t.Errorf("failed init logger %v", err)
	}

	handler := Newhandler(mockSt, &config.Config{}, logger)

	// Тестовые данные
	testUser := dataUser{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(testUser)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	// Выполняем обработчик
	handler.Register(rec, req)

	// Проверяем результат
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !mockSt.createUserCalled {
		t.Error("CreateUser was not called")
	}
}

func TestHandler_Login(t *testing.T) {
	mockSt := &mockStorage{
		user: &models.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Password: hashPassword(t, "password123"),
		},
	}

	// Создаём логгер
	logger, err := logger.Initialize()
	if err != nil {
		t.Errorf("failed init logger %v", err)
	}
	handler := Newhandler(mockSt, &config.Config{}, logger)

	testUser := dataUser{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(testUser)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Token  string    `json:"token"`
		UserId uuid.UUID `json:"user_id"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response.UserId != mockSt.user.ID {
		t.Errorf("expected user ID %s, got %s", mockSt.user.ID, response.UserId)
	}
}

func TestHandler_GetUserData(t *testing.T) {
	userID := uuid.New()
	mockSt := &mockStorage{
		userData: []*models.UserData{
			{
				ID: uuid.New(),
				// UserID:   userID,
				DataType: "password",
				Data:     []byte("encrypted-data"),
				Version:  1,
			},
		},
	}

	logger, err := logger.Initialize()
	if err != nil {
		t.Errorf("failed init logger %v", err)
	}
	handler := Newhandler(mockSt, &config.Config{}, logger)

	// Создаём контекст с userID
	ctx := context.WithValue(context.Background(), userId, userID)
	req := httptest.NewRequest(http.MethodGet, "/data", nil).WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetUserData(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response []*models.UserData
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 data item, got %d", len(response))
	}
}

func TestHandler_CreateData(t *testing.T) {
	userID := uuid.New()
	mockSt := &mockStorage{}

	logger, err := logger.Initialize()
	if err != nil {
		t.Errorf("failed init logger %v", err)
	}
	handler := Newhandler(mockSt, &config.Config{}, logger)

	requestData := struct {
		DataType string `json:"data_type"`
		Data     []byte `json:"encrypted_data"`
		Metadata []byte `json:"metadata"`
	}{
		DataType: "password",
		Data:     []byte("encrypted-password-data"),
		Metadata: []byte("metadata-for-password"),
	}

	jsonData, _ := json.Marshal(requestData)

	// Создаём контекст с userID
	ctx := context.WithValue(context.Background(), userId, userID)
	req := httptest.NewRequest(http.MethodPost, "/data", bytes.NewBuffer(jsonData)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateData(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response uuid.UUID
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if len(mockSt.userData) != 1 {
		t.Errorf("expected 1 data item in storage, got %d", len(mockSt.userData))
	}

	createdData := mockSt.userData[0]
	if createdData.DataType != requestData.DataType {
		t.Errorf("expected data type %s, got %s", requestData.DataType, createdData.DataType)
	}
}

// Вспомогательная функция для хеширования пароля в тестах
func hashPassword(t *testing.T, password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hashed)
}

func TestGetUserIDContx(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name    string
		ctx     context.Context
		key     string
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "valid context with user ID",
			ctx:     context.WithValue(context.Background(), userId, testUserID),
			key:     userId,
			wantID:  testUserID,
			wantErr: false,
		},
		{
			name:    "context without user ID",
			ctx:     context.Background(),
			key:     userId,
			wantID:  uuid.UUID{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := getUserIDContx(tt.ctx, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserIDContx() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotID != tt.wantID {
				t.Errorf("getUserIDContx() = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}
