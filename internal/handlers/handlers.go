// Package handlers содержит HTTP‑обработчики для эндпоинтов приложения.
package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/amfib87/go_musthave-GophKeeper/internal/crypto"
	"github.com/amfib87/go_musthave-GophKeeper/internal/logger"
	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/amfib87/go_musthave-GophKeeper/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	storage storage.Storage
	cfg     *config.Config
	logger  logger.Tlog
}

// Newhandler создаёт новый экземпляр обработчика с заданными зависимостями.
// Параметры:
//   - st (storage.Storage): интерфейс хранилища данных;
//   - cfg (*config.Config): конфигурация приложения;
//   - log (*logger.Tlog): экземпляр логгера.
//
// Возвращает *Handler — указатель на созданный обработчик.
func Newhandler(st storage.Storage, cfg *config.Config, log logger.Tlog) *Handler {
	return &Handler{
		storage: st,
		cfg:     cfg,
		logger:  log}
}

// Close освобождает ресурсы, связанные с обработчиком (например, соединение с БД).
func (h *Handler) Close() {
	defer h.logger.Log.Sync()
	defer h.storage.Close()
}

type dataUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const userId string = "user_id"

// Register обрабатывает регистрацию нового пользователя
func (h *Handler) Register(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started Register")
	var dataUser dataUser

	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		h.logger.Log.Error("failed buf.ReadFrom %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &dataUser); err != nil {
		h.logger.Log.Error("failed json.Unmarshal: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dataUser.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Log.Error("failed bcrypt.GenerateFromPassword: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    dataUser.Email,
		Password: string(passwordHash),
	}

	if err := h.storage.CreateUser(req.Context(), user); err != nil {
		h.logger.Log.Error("failed CreateUser %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		h.logger.Log.Error("failed generate JWT %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Генерируем JWT-токен
	var dataToken struct {
		Token  string    `json:"token"`
		UserId uuid.UUID `json:"user_id"`
	}

	dataToken.Token = token
	dataToken.UserId = user.ID

	resp, err := json.Marshal(dataToken)
	if err != nil {
		h.logger.Log.Error("failed Marshal %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	_, err = res.Write(resp)
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// Login обрабатывает аутентификацию пользователя
func (h *Handler) Login(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started Login")

	var dataUser dataUser

	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		h.logger.Log.Error("failed buf.ReadFrom %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &dataUser); err != nil {
		h.logger.Log.Error("failed json.Unmarshal: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByEmail(req.Context(), dataUser.Email)
	if err != nil || user == nil {
		h.logger.Log.Error("failed GetUserByEmail %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dataUser.Password)); err != nil {
		h.logger.Log.Error("the password does not match %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		h.logger.Log.Error("failed generate JWT %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Генерируем JWT-токен
	var dataToken struct {
		Token  string    `json:"token"`
		UserId uuid.UUID `json:"user_id"`
	}

	dataToken.Token = token
	dataToken.UserId = user.ID

	resp, err := json.Marshal(dataToken)
	if err != nil {
		h.logger.Log.Error("failed Marshal %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	_, err = res.Write(resp)
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// AuthMiddleware обеспечивает авторизацию запросов
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		h.logger.Log.Debug("started AuthMiddleware")

		token := req.Header.Get("Authorization")
		if token == "" {
			h.logger.Log.Error("token is empty")
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userID, err := parseJWT(token)
		if err != nil {
			h.logger.Log.Error("Invalid token", zap.String("token", token), zap.Any("userID", userID))
			http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), userId, userID)
		if next == nil {
			h.logger.Log.Error("failed context.WithValue %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func (h *Handler) GetUserData(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started GetUserData")

	userID, err := getUserIDContx(req.Context(), userId)
	if err != nil {
		h.logger.Log.Error("failed getUserIDContx %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	h.logger.Log.Debug("userID", zap.Any("userID", userID))
	data, err := h.storage.GetUserData(req.Context(), userID)
	if err != nil {
		h.logger.Log.Error("failed storage.GetUserData: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.logger.Log.Debug("data", zap.Any("data", data))
	resp, err := json.Marshal(data)
	if err != nil {
		h.logger.Log.Error("failed Marshal %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func getUserIDContx(cont context.Context, key string) (userID uuid.UUID, err error) {

	valUserID := cont.Value(key)
	if valUserID != nil {
		userID = valUserID.(uuid.UUID)
	} else {
		return userID, fmt.Errorf("userID is empty")
	}

	return userID, nil
}

func (h *Handler) CreateData(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started CreateData")

	userID, err := getUserIDContx(req.Context(), userId)
	if err != nil {
		h.logger.Log.Error("failed getUserIDContx %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var request struct {
		DataType string `json:"data_type"`
		Data     []byte `json:"encrypted_data"`
		Metadata string `json:"metadata"`
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(req.Body)
	if err != nil {
		h.logger.Log.Error("failed buf.ReadFrom %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &request); err != nil {
		h.logger.Log.Error("failed json.Unmarshal: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Шифруем данные перед сохранением
	encryptedData, err := crypto.EncryptData([]byte(request.Data), []byte(models.MasterPassword))

	userData := &models.UserData{
		ID:        uuid.New(),
		UserID:    userID,
		DataType:  request.DataType,
		Data:      encryptedData,
		Metadata:  request.Metadata,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.storage.CreateUserData(req.Context(), userData); err != nil {
		h.logger.Log.Error("storage.CreateUserData %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userData.ID)
	if err != nil {
		h.logger.Log.Error("failed Marshal %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(resp)
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteData(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started DeleteData")

	id, err := getUserIDContx(req.Context(), userId)
	if err != nil {
		h.logger.Log.Error("failed getUserIDContx %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	h.logger.Log.Debug("id", zap.Any("id", id))
	dataSl, err := h.storage.GetUserAllDataByID(req.Context(), id)
	if err != nil || dataSl == nil {
		h.logger.Log.Error("failed GetUserAllDataByID %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	for _, line := range dataSl {
		// Soft delete — помечаем как удалённое
		line.IsDeleted = true
		now := time.Now()
		line.DeletedAt = &now

		if err := h.storage.UpdateUserData(req.Context(), line); err != nil {
			h.logger.Log.Error("failed UpdateUserData %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusNoContent)
	_, err = res.Write([]byte(""))
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateData(res http.ResponseWriter, req *http.Request) {
	h.logger.Log.Debug("started UpdateData")

	userID, err := getUserIDContx(req.Context(), userId)
	if err != nil {
		h.logger.Log.Error("failed getUserIDContx %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var newData bool
	// Получаем существующие данные
	h.logger.Log.Debug("userID", zap.Any("userID", userID))

	existingData, err := h.storage.GetUserDataByID(req.Context(), userID)
	if !errors.Is(err, sql.ErrNoRows) {
		if err != nil {
			h.logger.Log.Error("failed GetUserDataByID %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		existingData = &models.UserData{
			UserID:    userID,
			CreatedAt: time.Now()}

	} else {
		existingData = &models.UserData{
			ID:        uuid.New(),
			UserID:    userID,
			UpdatedAt: time.Now()}
		newData = true
	}
	h.logger.Log.Debug("existingData", zap.Any("existingData", existingData))

	// Парсим входящие данные
	var request struct {
		DataType string `json:"data_type"`
		Data     []byte `json:"encrypted_data"`
		Metadata string `json:"metadata"`
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(req.Body)
	if err != nil {
		h.logger.Log.Error("failed buf.ReadFrom %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &request); err != nil {
		h.logger.Log.Error("failed json.Unmarshal: %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Обновляем данные
	existingData.DataType = request.DataType
	existingData.Data = request.Data
	existingData.Metadata = request.Metadata
	existingData.Version++

	if newData == false {
		h.logger.Log.Debug("UpdateUserData")
		if err := h.storage.UpdateUserData(req.Context(), existingData); err != nil {
			h.logger.Log.Error("failed UpdateUserData", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		h.logger.Log.Debug("CreateUserData")
		if err := h.storage.CreateUserData(req.Context(), existingData); err != nil {
			h.logger.Log.Error("storage.CreateUserData %v", zap.Error(err))
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	var answer struct {
		Id       uuid.UUID `json:"id"`
		Version  int       `json:"version"`
		UpdateAt time.Time `json:"updated_at"`
	}

	answer.Id = existingData.ID
	answer.Version = existingData.Version
	answer.UpdateAt = existingData.UpdatedAt

	resp, err := json.Marshal(answer)
	if err != nil {
		h.logger.Log.Error("failed Marshal %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)
	if err != nil {
		h.logger.Log.Error("failed res.Write %v", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
