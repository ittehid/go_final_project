package database

import (
	"database/sql"
	"fmt"
	"go_final_project/internal/logger"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() error {
	dbFile, err := getDBPath()
	if err != nil {
		logger.LogMessage("database", fmt.Sprintf("[ERROR] Не удалось определить путь к БД: %v", err))
	}
	logger.LogMessage("database", fmt.Sprintf("[INFO] Расположение базы данных: %s", dbFile))

	_, err = os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("не удалось открыть базу данных: %v", err)
	}

	DB = db
	if install {
		if err := createTables(); err != nil {
			db.Close()
			return err
		}
		logger.LogMessage("database", "[INFO] База данных создана.")
	}

	return nil
}

func getDBPath() (string, error) {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		return envDBFile, nil
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("не удалось получить рабочую директорию: %v", err)
	}

	dataDir := filepath.Join(projectDir, "data")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("не удалось создать папку data: %v", err)
	}

	return filepath.Join(dataDir, "scheduler.db"), nil
}

func createTables() error {
	const query = `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT CHECK(length(repeat) <= 128)
	);
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
	if _, err := DB.Exec(query); err != nil {
		return logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка создания таблицы: %v", err))
	}
	return nil
}

func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка при закрытии базы данных: %v", err))
		}
	}
}
