package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/google/uuid"
)

// HTTPClient представляет HTTP‑клиент для взаимодействия с сервером GophKeeper.
// Содержит конфигурацию подключения и токен аутентификации.
type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewHTTPClient создаёт новый экземпляр HTTP‑клиента.
// Параметры:
//
//	baseURL — базовый URL сервера API.
//
// Возвращает:
//
//	*HTTPClient — инициализированный клиент.
func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		baseURL: cfg.ServRunAddr,
		client:  &http.Client{},
	}
}

// GetUserData получает все данные пользователя с сервера.
// Возвращает:
//
//	[]*models.UserData — список данных пользователя.
//	error — ошибка при выполнении запроса.
func (c *HTTPClient) GetUserData() ([]*models.UserData, error) {

	req, err := http.NewRequest("GET", c.baseURL+"/api/data", nil)
	if err != nil {
		return nil, err
	}

	c.token, err = config.LoadAuthToken()
	if err != nil {
		fmt.Printf("Failed to load token: %v\n", err)
	}

	// Добавляем заголовок авторизации, если токен есть
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	} else {
		// Логируем отсутствие токена для отладки
		fmt.Println("Warning: Authorization token is empty")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []*models.UserData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

// RegisterUser регистрирует нового пользователя в системе.
// Параметры:
//
//	username — имя пользователя.
//	password — пароль пользователя (будет хеширован).
//
// Возвращает:
//
//	error — ошибка при регистрации (nil при успехе).
func (c *HTTPClient) RegisterUser(email, password string) (string, error) {
	requestBody := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error: %s", resp.Status)
	}

	// Парсим ответ с токеном
	var response struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Token, nil

}

// DeleteData отправляет запрос на удаление данных
func (c *HTTPClient) DeleteData(id string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/data/", nil)
	if err != nil {
		return err
	}

	c.token, err = config.LoadAuthToken()
	if err != nil {
		fmt.Printf("Failed to load token: %v\n", err)
	}

	// Добавляем заголовок авторизации, если токен есть
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	} else {
		// Логируем отсутствие токена для отладки
		fmt.Println("Warning: Authorization token is empty")
	}

	req.Header.Set("Authorization", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	return nil
}

// UpdateData отправляет запрос на обновление данных
func (c *HTTPClient) UpdateData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.baseURL+"/api/data/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	return nil
}

// LoginUser отправляет запрос на аутентификацию пользователя
func (c *HTTPClient) LoginUser(email, password string) (string, error) {
	requestBody := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error: %s", resp.Status)
	}

	// Парсим ответ с токеном
	var response struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	fmt.Println("token", response.Token)
	return response.Token, nil
}

// GetDataByID получает данные пользователя по ID с сервера
func (c *HTTPClient) GetDataByID(id uuid.UUID) (*models.UserData, error) {
	fmt.Println("GetDataByID")

	// Формируем URL с ID
	url := fmt.Sprintf("%s/api/data/%s", c.baseURL, id.String())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.token, err = config.LoadAuthToken()
	if err != nil {
		fmt.Printf("Failed to load token: %v\n", err)
	}

	// Добавляем заголовок авторизации, если токен есть
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	} else {
		// Логируем отсутствие токена для отладки
		fmt.Println("Warning: Authorization token is empty")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %s", resp.Status)
	}

	// Парсим ответ
	var response struct {
		Data *models.UserData `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// CreateUserData отправляет новые данные на сервер.
// Параметр:
//
//	data — данные для создания.
//
// Возвращает:
//
//	error — ошибка при выполнении запроса (nil при успехе).
func (c *HTTPClient) CreateUserData(data *models.UserData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/data", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Добавляем заголовок авторизации, если токен есть
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	} else {
		// Логируем отсутствие токена для отладки
		fmt.Println("Warning: Authorization token is empty")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}
	return nil
}

// UpdateUserData обновляет существующие данные на сервере.
// Параметр:
//
//	data — обновлённые данные.
//
// Возвращает:
//
//	error — ошибка при выполнении запроса (nil при успехе).
func (c *HTTPClient) UpdateUserData(data *models.UserData) error {

	url := fmt.Sprintf("%s/api/data/%s", c.baseURL, data.ID.String())
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}
	return nil
}
