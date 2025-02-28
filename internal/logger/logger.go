package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logDir = "logs"

func init() {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("[ERROR] Папка для логов не создана: %v", err)
	}
}

func getLogFilePath(logsType string) string {
	date := time.Now().Format("02-01-2006")
	fileName := fmt.Sprintf("%s_%s.log", logsType, date)

	return filepath.Join(logDir, fileName)
}

func LogMessage(logsType, message string) error {
	logFilePath := getLogFilePath(logsType)

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("[ERROR] Не удалось открыть файл лога: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("02.01.2006 15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	logger := log.New(file, "", 0)
	logger.Println(formattedMessage)

	return nil
}
