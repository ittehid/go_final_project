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
	return fmt.Sprintf("%d", config.Port)
}

func runServer(port string) error {
	db := database.GetDB()

	http.HandleFunc("/api/nextdate", scheduler.NextDateHandler())

	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			logger.LogMessage("api", "[INFO] POST-запрос на добавление задачи")
			task.AddTaskHandler(db)(w, r)
		case http.MethodGet:
			logger.LogMessage("api", "[INFO] GET-запрос на получение задачи")
			task.GetTaskHandler(db)(w, r)
		case http.MethodPut:
			logger.LogMessage("api", "[INFO] PUT-запрос на изменение задачи")
			task.EditTaskHandler(db)(w, r)
		case http.MethodDelete:
			logger.LogMessage("api", "[INFO] DELETE-запрос на завершение задачи")
			task.DoneTaskHandler(db)(w, r)
		default:
			logger.LogMessage("api", fmt.Sprintf("[ERROR] Неподдерживаемый метод запроса: %s", r.Method))
			http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/task/done", scheduler.AuthMiddleware(task.DoneTaskHandler(db)))
	http.HandleFunc("/api/tasks", scheduler.AuthMiddleware(task.GetTasksHandler(db)))

	http.Handle("/", http.FileServer(http.Dir("web")))

	http.HandleFunc("/api/signin", scheduler.SignInHandler)

	logger.LogMessage("server", "[INFO] Обработчики запросов успешно зарегистрированы")
	return http.ListenAndServe(":"+port, nil)
}
