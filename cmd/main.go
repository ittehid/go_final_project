package main

import (
	"fmt"
	"net/http"
	"os"

	"go_final_project/internal/database"
	"go_final_project/internal/logger"
	"go_final_project/internal/scheduler"
	"go_final_project/internal/task"
	"go_final_project/tests"
)

func main() {
	port := getPort()

	logger.LogMessage("server", "[INFO] Запуск инициализации базы данных")
	if err := database.InitDB(); err != nil {
		logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка инициализации базы данных: %v", err))
		return
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
	return fmt.Sprintf("%d", tests.Port)
}

func runServer(port string) error {
	db := database.GetDB()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/task", task.AddTaskHandler(db))
	mux.HandleFunc("GET /api/task", task.GetTaskHandler(db))
	mux.HandleFunc("PUT /api/task", task.EditTaskHandler(db))
	mux.HandleFunc("DELETE /api/task", task.DoneTaskHandler(db))

	mux.HandleFunc("/api/nextdate", scheduler.NextDateHandler())
	mux.Handle("/api/task/done", scheduler.AuthMiddleware(task.DoneTaskHandler(db)))
	mux.Handle("/api/tasks", scheduler.AuthMiddleware(task.GetTasksHandler(db)))

	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.HandleFunc("/api/signin", scheduler.SignInHandler)

	logger.LogMessage("server", "[INFO] Обработчики запросов успешно зарегистрированы")
	return http.ListenAndServe(":"+port, mux)
}
