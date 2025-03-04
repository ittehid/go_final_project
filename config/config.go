package config

import (
	"go_final_project/internal/logger"
	"os"
	"path/filepath"
	"runtime"
)

func GetDBFilePath() string {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		logger.LogMessage("config", "[INFO] Используется путь к БД из переменной окружения: "+envDBFile)
		return envDBFile
	}

	// Определяем корень проекта на основе пути к этому файлу
	_, filename, _, _ := runtime.Caller(0)
	projectDir := filepath.Dir(filepath.Dir(filename)) // Поднимаемся на уровень выше (из `config` в корень проекта)

	dataDir := filepath.Join(projectDir, "data")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		logger.LogMessage("config", "[ERROR] Ошибка при создании папки data: "+err.Error())
		return "scheduler.db"
	}

	dbPath := filepath.Join(dataDir, "scheduler.db")
	logger.LogMessage("config", "[INFO] Путь к базе данных: "+dbPath)
	return dbPath
}
