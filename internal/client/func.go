package client

import (
	"fmt"

	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/amfib87/go_musthave-GophKeeper/internal/crypto"
	"github.com/amfib87/go_musthave-GophKeeper/internal/models"
	"github.com/amfib87/go_musthave-GophKeeper/internal/storage"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var version = "1.0.0"
var buildDate = "2026-05-24"

type SyncResult struct {
	PushedToServer    int
	PulledFromServer  int
	MergedLocally     int
	ConflictsResolved int
	MergedData        []*models.UserData
}

func LoginCmd(cl *HTTPClient) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to GophKeeper",
		Run: func(cmd *cobra.Command, args []string) {
			if email == "" || password == "" {
				fmt.Println("Please provide both email and password")
				return
			}

			// Отправляем запрос на аутентификацию
			token, err := cl.LoginUser(email, password)
			if err != nil {
				fmt.Printf("Login failed: %v\n", err)
				return
			}

			// Сохраняем токен в конфигурационный файл
			cl.token = token

			// Сохраняем токен на диск
			if err := config.SaveAuthToken(token); err != nil {
				fmt.Printf("Warning: failed to save token: %v\n", err)
			}

			fmt.Println("Login successful!")
			fmt.Println("Authentication token saved.")
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User email (required)")
	cmd.Flags().StringVar(&password, "password", "", "User password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func RegisterCmd(cl *HTTPClient) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new user",
		Run: func(cmd *cobra.Command, args []string) {
			if email == "" || password == "" {
				fmt.Println("Please provide both email and password")
				return
			}

			// Отправляем запрос на регистрацию
			token, err := cl.RegisterUser(email, password)
			if err != nil {
				fmt.Printf("Registration failed: %v\n", err)
				return
			}

			cl.token = token

			fmt.Println("User registered successfully!")
			fmt.Println("You can now login with your credentials.")
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User email (required)")
	cmd.Flags().StringVar(&password, "password", "", "User password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nBuild date: %s\n", version, buildDate)
		},
	}
}

func ListCmd(cl *HTTPClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stored data",
		Run: func(cmd *cobra.Command, args []string) {

			// Получаем данные с сервера
			data, err := cl.GetUserData()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Your stored data:")
			for _, item := range data {
				fmt.Printf("- ID: %s, Type: %s, Data: %s, Meatdata: %s\n", item.ID, item.DataType, string(item.Data), item.Metadata)
			}
		},
	}
}

// getCmd создаёт команду получения данных по ID для CLI.
// Возвращает:
//
//	*cobra.Command — команда получения данных.
func GetCmd(cl *HTTPClient) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get specific data by ID",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" {
				fmt.Println("Please provide data ID")
				return
			}

			// Преобразуем строку в UUID
			idUuid, err := uuid.Parse(id)
			if err != nil {
				fmt.Printf("Invalid ID format: %v\n", err)
				return
			}

			// Получаем данные с сервера
			data, err := cl.GetDataByID(idUuid)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Расшифровываем данные мастер‑паролем
			decryptedData, err := crypto.DecryptData(data.Data, []byte(cl.masterPassword))
			if err != nil {
				fmt.Printf("Decryption error: %v\n", err)
				return
			}

			fmt.Printf("Data: %s\n", string(decryptedData))
			fmt.Printf("Metadata: %s\n", string(data.Metadata))
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Data ID")
	return cmd
}

func DeleteCmd(cl *HTTPClient) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete data by ID",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" {
				fmt.Println("Please provide data ID")
				return
			}

			// Отправляем запрос на удаление
			err := cl.DeleteData(id)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Data deleted successfully!")
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Data ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func UpdateCmd(cl *HTTPClient) *cobra.Command {
	var dataType, metadata, data string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update existing user data",
		Run: func(cmd *cobra.Command, args []string) {

			// Загружаем токен из файла
			token, err := config.LoadAuthToken()
			if err != nil {
				fmt.Printf("Failed to load token: %v\n", err)
				return
			}
			cl.SetAuthToken(token)

			// Создаём объект данных для обновления
			userData := &models.UserData{
				DataType: dataType,
				Data:     []byte(data),
				Metadata: metadata,
			}

			// Вызываем функцию обновления данных
			err = cl.UpdateData(userData)
			if err != nil {
				fmt.Printf("Error updating data: %v\n", err)
				return
			}

			fmt.Println("Data created successfully!")
		},
	}

	// Добавляем флаги для ввода данных
	cmd.Flags().StringVar(&dataType, "type", "", "Type of data (password, note, etc.)")
	cmd.Flags().StringVar(&data, "data", "", "Data")
	cmd.Flags().StringVar(&metadata, "metadata", "", "Additional metadata as JSON")

	return cmd
}

// syncCmd создаёт команду синхронизации для CLI.
// Команда выполняет:
// 1. Загрузка данных с сервера.
// 2. Загрузка локальных данных.
// 3. Синхронизация данных.
// 4. Сохранение обновлённых данных.
// 5. Вывод отчёта.
// Возвращает:
//
//	*cobra.Command — команда синхронизации.
func SyncCmd(cl *HTTPClient) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {

			// Инициализируем локальное хранилище
			storage := storage.NewLocalStorage(config.GetDataDirectory())

			// Получаем последнюю версию данных с сервера
			serverData, err := cl.GetUserData()
			if err != nil {
				fmt.Printf("Sync error: %v\n", err)
				return
			}

			// Загружаем локальные данные
			localData, err := storage.GetLocalData()
			if err != nil {
				fmt.Printf("Failed to load local data: %v\n", err)
				return
			}

			// Выполняем синхронизацию
			syncResult, err := performSync(cl, localData, serverData)
			if err != nil {
				fmt.Printf("Synchronization failed: %v\n", err)
				return
			}

			// Сохраняем обновлённые локальные данные
			err = storage.SaveLocalData(syncResult.MergedData)
			if err != nil {
				fmt.Printf("Failed to save local data: %v\n", err)
				return
			}

			// Выводим отчёт о синхронизации
			printSyncReport(syncResult)
		},
	}
}

// printSyncReport выводит отчёт о синхронизации в читаемом формате.
// Параметр:
//
//	result — результат синхронизации для отображения.
func printSyncReport(result *SyncResult) {
	fmt.Println("=== Synchronization ===")
	fmt.Printf("✓ Synchronization completed successfully!\n")
	fmt.Printf("Pushed to server: %d entries\n", result.PushedToServer)
	fmt.Printf("Pulled from server: %d entries\n", result.PulledFromServer)
	fmt.Printf("Merged locally: %d entries\n", result.MergedLocally)
	fmt.Printf("Conflicts resolved: %d\n", result.ConflictsResolved)
	fmt.Println("Your data is now synchronized with the server.")
}

// performSync выполняет синхронизацию локальных и серверных данных.
// Алгоритм:
// 1. Сравнивает версии данных.
// 2. Загружает новые данные с сервера.
// 3. Отправляет новые данные на сервер.
// 4. Разрешает конфликты версий.
// Параметры:
//
//	cl — HTTP‑клиент для взаимодействия с сервером.
//	local — локальные данные.
//	server — данные с сервера.
//
// Возвращает:
//
//	*SyncResult — результат синхронизации.
//	error — ошибка при синхронизации (nil при успехе).
func performSync(cl *HTTPClient, local, server []*models.UserData) (*SyncResult, error) {
	result := &SyncResult{
		MergedData:       make([]*models.UserData, 0),
		PulledFromServer: 0,
		PushedToServer:   0,
	}

	localMap := make(map[uuid.UUID]*models.UserData)

	serverMap := make(map[uuid.UUID]*models.UserData)

	// Создаём мапы для быстрого поиска по ID
	for _, data := range local {
		localMap[data.ID] = data
	}
	for _, data := range server {
		serverMap[data.ID] = data
	}

	// Обрабатываем данные, которые есть на сервере
	for id, serverData := range serverMap {
		localData, exists := localMap[id]

		if !exists {
			// Данные есть только на сервере — скачиваем
			result.MergedData = append(result.MergedData, serverData)
			result.PulledFromServer++
		} else {
			// Данные есть и локально, и на сервере — проверяем версии
			if serverData.Version > localData.Version {
				// Серверная версия новее — используем её
				result.MergedData = append(result.MergedData, serverData)
				result.PulledFromServer++
			} else if serverData.Version < localData.Version {
				// Локальная версия новее — отправляем на сервер
				err := cl.UpdateUserData(localData)
				if err == nil {
					result.MergedData = append(result.MergedData, localData)
					result.PushedToServer++
				} else {
					fmt.Printf("Failed to update data %s on server: %v\n", localData.ID, err)
				}
			} else {
				// Версии совпадают — используем локальную
				result.MergedData = append(result.MergedData, localData)
			}
		}
	}

	// Добавляем локальные данные, которых нет на сервере (новые записи)
	for id, localData := range localMap {
		if _, exists := serverMap[id]; !exists {
			err := cl.CreateUserData(localData)
			if err == nil {
				result.MergedData = append(result.MergedData, localData)
				result.PushedToServer++
			} else {
				fmt.Printf("Failed to create data %s on server: %v\n", localData.ID, err)
			}
		}
	}

	return result, nil
}

func (c *HTTPClient) SetAuthToken(token string) {
	c.token = token
}
