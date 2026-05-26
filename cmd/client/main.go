// Package main — точка входа в клиентское приложение GophKeeper.
//
// GophKeeper CLI — консольный клиент для взаимодействия с сервером GophKeeper,
// предоставляющий функционал для:
//   - регистрации и авторизации пользователей;
//   - управления пользовательскими данными (пароли, заметки, карты и т. д.);
//   - синхронизации данных с сервером.
//
// Использует библиотеку Cobra для построения CLI‑интерфейса.
package main

import (
	"fmt"
	"os"

	"github.com/amfib87/go_musthave-GophKeeper/internal/client"
	cl "github.com/amfib87/go_musthave-GophKeeper/internal/client"
	"github.com/amfib87/go_musthave-GophKeeper/internal/config"
	"github.com/spf13/cobra"
)

// main — точка входа в приложение.
// Выполняет следующие действия:
// 1. Создаёт и парсит конфигурацию приложения (NewConfig + ParseFlag).
// 2. Инициализирует корневой командный узел Cobra.
// 3. Добавляет все доступные команды CLI.
// 4. Запускает выполнение команд через Cobra.
//
// При возникновении ошибки выводит сообщение и завершает программу с кодом 1.
func main() {

	// Парсим флаги
	cfg := config.NewConfig()
	cfg.ParseFlag()

	rootCmd := &cobra.Command{Use: "gophkeeper"}

	// Добавляем глобальный флаг -a (адрес сервера)
	rootCmd.PersistentFlags().StringVar(
		&cfg.ServRunAddr,
		"a",
		"http://localhost:8080",
		"server address for connection",
	)

	// Инициализируем HTTP‑клиент
	client := client.NewHTTPClient(cfg)

	rootCmd.AddCommand(
		cl.RegisterCmd(client),
		cl.LoginCmd(client),

		cl.ListCmd(client),
		// cl.GetCmd(client),
		cl.DeleteCmd(client),
		cl.SyncCmd(client),
		cl.VersionCmd(),
		cl.UpdateCmd(client),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
