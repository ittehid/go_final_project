package database

import (
	"fmt"
	"os"

	"go_final_project/config"
	"go_final_project/internal/logger"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func InitDB() error {
	dbFile := config.GetDBFilePath()
	logger.LogMessage("database", fmt.Sprintf("[INFO] Расположение базы данных: %s", dbFile))

	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sqlx.Open("sqlite", dbFile)
	if err != nil {
		logger.LogMessage("database", fmt.Sprintf("[ERROR] Не удалось открыть базу данных: %v", err))
		return fmt.Errorf("не удалось открыть базу данных: %v", err)
	}

	DB = db
	if install {
		logger.LogMessage("database", "[INFO] База данных не найдена, начинаем создание таблиц")
		if err := createTables(); err != nil {
			logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка при создании таблиц: %v", err))
			db.Close()
			return err
		}
		logger.LogMessage("database", "[INFO] База данных создана.")
	}
	logger.LogMessage("database", "[INFO] Инициализация базы данных завершена успешно")
	return nil
}

func createTables() error {
	const query = `
	CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT CHECK(length(repeat) <= 128)
	);
	CREATE INDEX idx_date ON scheduler (date);
	`
	_, err := DB.Exec(query)
	if err != nil {
		logger.LogMessage("database", fmt.Sprintf("[ERROR] Ошибка создания таблицы: %v", err))
		return fmt.Errorf("ошибка создания таблицы: %v", err)
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

func GetDB() *sqlx.DB {
	return DB
}
