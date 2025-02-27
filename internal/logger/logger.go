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
		log.Fatalf("[ERROR] Директории логов не создана: %v", err)
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

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(message)
	return nil
}
