package scheduler

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

func NextDateHandler(dbase *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		nowStr := req.URL.Query().Get("now")
		dateStr := req.URL.Query().Get("date")
		repeatStr := req.URL.Query().Get("repeat")

		now, err := parseNow(nowStr)
		if err != nil {
			http.Error(w, "invalid 'now' parameter", http.StatusBadRequest)
			return
		}

		if _, err := time.Parse("20060102", dateStr); err != nil {
			http.Error(w, "invalid 'date' parameter", http.StatusBadRequest)
			return
		}

		nextDate, err := NextDate(now, dateStr, repeatStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(nextDate))
	}
}

func parseNow(nowStr string) (time.Time, error) {
	if nowStr == "" {
		return time.Now(), nil
	}
	return time.Parse("20060102", nowStr)
}
