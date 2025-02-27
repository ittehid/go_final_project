package main

import (
	"fmt"
	"go_final_project/config"
	"go_final_project/internal/logger"
	"log"
	"net/http"
	"os"
)

func main() {
	port := getPort()
	logger.LogMessage("server", fmt.Sprintf("[INFO] Сервер запущен. Порт: %s", port))
	log.Printf("[INFO] Сервер запущен. Порт: %s", port)

	if err := runServer(port); err != nil {
		logger.LogMessage("server", fmt.Sprintf("[ERROR] Ошибка запуска сервера: %v", err))
		log.Fatalf("[ERROR] Ошибка запуска сервера: %v", err)
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
	http.Handle("/", http.FileServer(http.Dir("web")))
	return http.ListenAndServe(":"+port, nil)
}
