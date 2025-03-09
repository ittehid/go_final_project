package scheduler

import (
	"fmt"
	"net/http"
	"time"

	"go_final_project/internal"
	"go_final_project/internal/logger"
)

func NextDateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		nowStr := req.URL.Query().Get("now")
		dateStr := req.URL.Query().Get("date")
		repeatStr := req.URL.Query().Get("repeat")

		now, err := parseNow(nowStr)
		if err != nil {
			logger.LogMessage("scheduler", fmt.Sprintf("[ERROR] Некорректный параметр 'now': %v", err))
			http.Error(w, "некорректный параметр 'now'", http.StatusBadRequest)
			return
		}

		if _, err := time.Parse(internal.DateLayout, dateStr); err != nil {
			logger.LogMessage("scheduler", fmt.Sprintf("[ERROR] Некорректный параметр 'date': %v", err))
			http.Error(w, "некорректный параметр 'date'", http.StatusBadRequest)
			return
		}

		nextDate, err := NextDate(now, dateStr, repeatStr)
		if err != nil {
			logger.LogMessage("scheduler", fmt.Sprintf("[ERROR] Ошибка вычисления следующей даты: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.LogMessage("scheduler", fmt.Sprintf("[INFO] Успешно рассчитана следующая дата: %s", nextDate))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(nextDate))
	}
}

func parseNow(nowStr string) (time.Time, error) {
	if nowStr == "" {
		return time.Now(), nil
	}
	return time.Parse(internal.DateLayout, nowStr)
}
