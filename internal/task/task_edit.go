package task

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_final_project/internal/scheduler"

	"github.com/jmoiron/sqlx"
)

// GetTaskHandler обрабатывает GET-запрос для получения задачи по ID
func GetTaskHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
			return
		}

		task, err := getTaskByID(db, id)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

// EditTaskHandler обрабатывает PUT-запрос для редактирования задачи
func EditTaskHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
			return
		}

		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, `{"error":"ошибка разбора JSON"}`, http.StatusBadRequest)
			return
		}

		if task.ID == "" {
			http.Error(w, `{"error":"не указан идентификатор задачи"}`, http.StatusBadRequest)
			return
		}

		// Проверка валидности ID
		if _, err := strconv.ParseInt(task.ID, 10, 64); err != nil {
			http.Error(w, `{"error":"некорректный идентификатор задачи"}`, http.StatusBadRequest)
			return
		}

		// Проверка наличия задачи в БД
		if _, err := getTaskByID(db, task.ID); err != nil {
			http.Error(w, `{"error":"задача не найдена"}`, http.StatusNotFound)
			return
		}

		// Валидация задачи (проверка даты и заголовка)
		if err := task.Validate(); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		// Дополнительная проверка корректности Repeat (если указано)
		if task.Repeat != "" {
			today := time.Now()
			if _, err := scheduler.NextDate(today, task.Date, task.Repeat); err != nil {
				http.Error(w, `{"error":"некорректное правило повторения"}`, http.StatusBadRequest)
				return
			}
		}

		// Обновление задачи
		if err := updateTask(db, &task); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}
}

// getTaskByID ищет задачу по идентификатору
func getTaskByID(db *sqlx.DB, id string) (*Task, error) {
	var task Task
	var numericID int64
	var err error

	numericID, err = strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Printf("Ошибка преобразования ID задачи (%s): %v", id, err)
		return nil, errors.New("Некорректный идентификатор задачи")
	}

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err = db.Get(&task, query, numericID)
	if err != nil {
		log.Printf("Задача с ID %d не найдена: %v", numericID, err)
		return nil, errors.New("Задача не найдена")
	}

	task.ID = strconv.FormatInt(numericID, 10)
	return &task, nil
}

// updateTask обновляет задачу в БД
func updateTask(db *sqlx.DB, task *Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	_, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("Ошибка обновления задачи с ID %s: %v", task.ID, err)
		return errors.New("ошибка обновления задачи")
	}
	return nil
}
