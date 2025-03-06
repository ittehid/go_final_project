package task

import (
	"encoding/json"
	"errors"
	"go_final_project/internal/scheduler"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// Task — структура задачи
type Task struct {
	ID      string `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment,omitempty"`
	Repeat  string `db:"repeat" json:"repeat,omitempty"`
}

// AddTaskHandler обрабатывает HTTP-запросы для добавления задач
func AddTaskHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			addTask(w, r, db)
		default:
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	}
}

// addTask обрабатывает POST-запрос для добавления задачи
func addTask(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error":"ошибка разбора JSON"}`, http.StatusBadRequest)
		return
	}

	// Валидируем поля
	if err := task.Validate(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Корректируем дату
	if err := task.AdjustDate(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Сохраняем в БД
	id, err := task.Save(db)
	if err != nil {
		log.Println("Ошибка сохранения в БД:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

// Validate проверяет корректность полей
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("не указан заголовок задачи")
	}
	if t.Date == "" {
		t.Date = time.Now().Format("20060102")
	}
	if _, err := time.Parse("20060102", t.Date); err != nil {
		return errors.New("дата указана в неверном формате YYYYMMDD")
	}
	return nil
}

// AdjustDate корректирует дату, если она в прошлом
func (t *Task) AdjustDate() error {
	todayStr := time.Now().Format("20060102")

	// Проверяем, является ли дата задачи прошлой
	if t.Date < todayStr {
		if t.Repeat == "" {
			t.Date = todayStr
		} else {
			currentDate, err := time.Parse("20060102", todayStr)
			if err != nil {
				return errors.New("ошибка обработки даты")
			}

			nextDate, err := scheduler.NextDate(currentDate, t.Date, t.Repeat)
			if err != nil {
				return errors.New("ошибка в правиле повторения")
			}
			t.Date = nextDate
		}
	}

	return nil
}

// Save сохраняет задачу в БД
func (t *Task) Save(db *sqlx.DB) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		log.Println("Ошибка SQL:", err)
		return 0, errors.New("ошибка сохранения в БД")
	}
	return res.LastInsertId()
}

// GetTasksHandler обрабатывает HTTP-запросы для получения задач
func GetTasksHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTasks(w, r, db)
		default:
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	}
}

// getTasks обрабатывает GET-запрос для получения задач
func getTasks(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	// Получаем параметр search из URL
	search := r.URL.Query().Get("search")
	limit := 50 // Лимит на количество задач

	var query string
	var args []interface{}

	// Если search соответствует формату даты, то ищем задачи по дате
	if isValidDateFormat(search) {
		dateFilter := convertToDBDateFormat(search)
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?"
		args = append(args, dateFilter, limit)
	} else if search != "" {
		// Иначе ищем по строкам title и comment
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?"
		args = append(args, "%"+search+"%", "%"+search+"%", limit)
	} else {
		// Если search пустой, возвращаем все задачи, отсортированные по дате
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?"
		args = append(args, limit)
	}

	// Выполняем запрос к базе данных
	var tasks []Task
	err := db.Select(&tasks, query, args...)
	if err != nil {
		http.Error(w, `{"error":"ошибка при извлечении данных"}`, http.StatusInternalServerError)
		return
	}

	// Формируем ответ в формате JSON
	w.Header().Set("Content-Type", "application/json")
	if len(tasks) == 0 {
		tasks = []Task{} // Чтобы не вернуть null в JSON
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	json.NewEncoder(w).Encode(response)
}

// isValidDateFormat проверяет, соответствует ли строка формату даты "DD.MM.YYYY"
func isValidDateFormat(dateStr string) bool {
	_, err := time.Parse("02.01.2006", dateStr)
	return err == nil
}

// convertToDBDateFormat преобразует дату в формат "YYYYMMDD", который используется в базе
func convertToDBDateFormat(dateStr string) string {
	t, _ := time.Parse("02.01.2006", dateStr)
	return t.Format("20060102")
}
