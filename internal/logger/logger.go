package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logDir = "logs"

var (
	logFile    *os.File
	logger     *log.Logger
	logChannel chan string
)

func init() {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("[ERROR] Папка для логов не создана: %v", err)
	}

	logFilePath := getLogFilePath("log")
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[ERROR] Не удалось открыть файл лога: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	logChannel = make(chan string, 100)

	go processLogs()
}

func getLogFilePath(logsType string) string {
	date := time.Now().Format("02-01-2006")
	fileName := fmt.Sprintf("%s_%s.log", logsType, date)

	return filepath.Join(logDir, fileName)
}

func processLogs() {
	for msg := range logChannel {
		logger.Println(msg)
	}
}

func LogMessage(logsType, message string) {
	timestamp := time.Now().Format("02.01.2006 15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	select {
	case logChannel <- formattedMessage:
	default:
		fmt.Println("[ERROR] Логгер перегружен, сообщение пропущено:", formattedMessage)
	}
}

func CloseLogger() {
	close(logChannel)
	logFile.Close()
}
