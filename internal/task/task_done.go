package task

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"go_final_project/internal"
	"go_final_project/internal/scheduler"

	"github.com/jmoiron/sqlx"
)

// DoneTaskHandler обрабатывает запросы для завершения и удаления задач
func DoneTaskHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		// Отдельная проверка для DELETE-запроса без ID
		if r.Method == http.MethodDelete && id == "" {
			http.Error(w, `{"error":"не указан идентификатор задачи"}`, http.StatusBadRequest)
			return
		}

		if id == "" {
			http.Error(w, `{"error":"не указан идентификатор задачи"}`, http.StatusBadRequest)
			return
		}

		task, err := getTaskByID(db, id)
		if err != nil {
			http.Error(w, `{"error":"задача не найдена"}`, http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodPost:
			if task.Repeat == "" {
				err = deleteTask(db, id)
				if err != nil {
					http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
					return
				}
			} else {
				today, _ := time.Parse(internal.DateLayout, task.Date)
				nextDate, err := scheduler.NextDate(today, task.Date, task.Repeat)
				if err != nil {
					http.Error(w, `{"error":"ошибка расчёта следующей даты"}`, http.StatusInternalServerError)
					return
				}
				err = updateTaskDate(db, id, nextDate)
				if err != nil {
					http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
					return
				}
			}

		case http.MethodDelete:
			err = deleteTask(db, id)
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
			return
		}

		// Успешный ответ
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}

// deleteTask удаляет задачу по ID
func deleteTask(db *sqlx.DB, id string) error {
	_, err := db.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		log.Printf("Ошибка удаления задачи с ID %s: %v", id, err)
		return errors.New("ошибка удаления задачи")
	}
	return nil
}

// updateTaskDate обновляет дату задачи
func updateTaskDate(db *sqlx.DB, id, date string) error {
	_, err := db.Exec("UPDATE scheduler SET date=? WHERE id=?", date, id)
	if err != nil {
		log.Printf("Ошибка обновления даты задачи с ID %s: %v", id, err)
		return errors.New("ошибка обновления даты задачи")
	}
	return nil
}
