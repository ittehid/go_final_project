package task

import (
	"encoding/json"
	"errors"
	"go_final_project/internal/logger"
	"log"
	"net/http"
	"time"

	"go_final_project/internal"
	"go_final_project/internal/scheduler"

	"github.com/jmoiron/sqlx"
)

// Task описывает задачу
type Task struct {
	ID      string `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment,omitempty"`
	Repeat  string `db:"repeat" json:"repeat,omitempty"`
}

type Repository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(t *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := r.db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		logger.LogMessage("[ERROR] Ошибка SQL: " + err.Error())
		log.Println("Ошибка SQL:", err)
		return 0, errors.New("ошибка сохранения в БД")
	}
	return res.LastInsertId()
}

func (t *Task) Validate() error {
	if t.Title == "" {
		logger.LogMessage("[ERROR] Не указан заголовок задачи")
		return errors.New("не указан заголовок задачи")
	}
	if t.Date == "" {
		t.Date = time.Now().Format(internal.DateLayout)
	}
	if _, err := time.Parse(internal.DateLayout, t.Date); err != nil {
		logger.LogMessage("[ERROR] Дата указана в неверном формате YYYYMMDD")
		return errors.New("дата указана в неверном формате YYYYMMDD")
	}
	return nil
}

func (t *Task) AdjustDate() error {
	todayStr := time.Now().Format(internal.DateLayout)
	// Проверяем, является ли дата задачи прошлой
	if t.Date < todayStr {
		// Если повторения нет, просто ставим сегодняшнюю дату
		if t.Repeat == "" {
			t.Date = todayStr
		} else {
			currentDate, err := time.Parse(internal.DateLayout, todayStr)
			if err != nil {
				logger.LogMessage("[ERROR] Ошибка обработки текущей даты")
				return errors.New("ошибка обработки текущей даты")
			}

			nextDate, err := scheduler.NextDate(currentDate, t.Date, t.Repeat)
			if err != nil {
				logger.LogMessage("[ERROR] Ошибка в правиле повторения")
				return errors.New("ошибка в правиле повторения")
			}
			t.Date = nextDate
		}
	}
	return nil
}

func AddTaskHandler(db *sqlx.DB) http.HandlerFunc {
	repo := NewTaskRepository(db)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			addTask(w, r, repo)
		default:
			logger.LogMessage("[ERROR] Метод не поддерживается")
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	}
}

func addTask(w http.ResponseWriter, r *http.Request, repo *Repository) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		logger.LogMessage("[ERROR] Ошибка разбора JSON")
		http.Error(w, `{"error":"ошибка разбора JSON"}`, http.StatusBadRequest)
		return
	}

	// Валидируем поля задачи
	if err := task.Validate(); err != nil {
		logger.LogMessage("[ERROR] " + err.Error())
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Корректируем дату, если нужно
	if err := task.AdjustDate(); err != nil {
		logger.LogMessage("[ERROR] " + err.Error())
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Сохраняем в БД (через репозиторий)
	id, err := repo.Save(&task)
	if err != nil {
		logger.LogMessage("[ERROR] Ошибка сохранения в БД")
		log.Println("Ошибка сохранения в БД:", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func GetTasksHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTasks(w, r, db)
		default:
			logger.LogMessage("[ERROR] Метод не поддерживается")
			http.Error(w, `{"error":"метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	}
}

func getTasks(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	search := r.URL.Query().Get("search")
	limit := internal.TaskLimit

	var query string
	var args []interface{}

	// Если search соответствует формату даты "DD.MM.YYYY"
	if isValidDateFormat(search) {
		dateFilter := convertToDBDateFormat(search)
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?"
		args = append(args, dateFilter, limit)
	} else if search != "" {
		// Ищем по строкам title и comment
		query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?"
		args = append(args, "%"+search+"%", "%"+search+"%", limit)
	} else {
		// Если search пустой, возвращаем все задачи, отсортированные по дате
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?"
		args = append(args, limit)
	}

	var tasks []Task
	err := db.Select(&tasks, query, args...)
	if err != nil {
		logger.LogMessage("[ERROR] Ошибка при извлечении данных")
		http.Error(w, `{"error":"ошибка при извлечении данных"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(tasks) == 0 {
		tasks = []Task{} // Чтобы в JSON был пустой массив, а не null
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	json.NewEncoder(w).Encode(response)
}

// isValidDateFormat проверяет, соответствует ли строка формату "DD.MM.YYYY"
func isValidDateFormat(dateStr string) bool {
	_, err := time.Parse(internal.DateFormatDDMMYYYY, dateStr)
	return err == nil
}

// convertToDBDateFormat преобразует "DD.MM.YYYY" -> "YYYYMMDD"
func convertToDBDateFormat(dateStr string) string {
	t, _ := time.Parse(internal.DateFormatDDMMYYYY, dateStr)
	return t.Format(internal.DateLayout)
}
