package main

import (
	"fmt"
	"go_final_project/config"
	"go_final_project/internal/database"
	"go_final_project/internal/logger"
	"go_final_project/internal/scheduler"
	"go_final_project/internal/task"
	"net/http"
	"os"
)

func main() {
	port := getPort()

	if err := database.InitDB(); err != nil {
		logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка: %v", err))
	}
	defer database.CloseDB()

	logger.LogMessage("server", fmt.Sprintf("[INFO] Сервер запущен. Порт: %s", port))

	if err := runServer(port); err != nil {
		logger.LogMessage("server", fmt.Sprintf("[ERROR] Ошибка запуска сервера: %v", err))
	}
}

func getPort() string {
	reqPort := os.Getenv("TODO_PORT")
	if reqPort != "" {
		return reqPort
	}
	return fmt.Sprintf("%d", config.Port)
}

func runServer(port string) error {
	db := database.GetDB()
	http.HandleFunc("/api/nextdate", scheduler.NextDateHandler())
	http.HandleFunc("/api/task", task.AddTaskHandler(db))
	http.Handle("/", http.FileServer(http.Dir("web")))
	return http.ListenAndServe(":"+port, nil)
}
