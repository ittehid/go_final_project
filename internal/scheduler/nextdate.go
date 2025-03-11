package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go_final_project/internal"
	"go_final_project/internal/logger"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		logger.LogMessage("[ERROR] Повтор пуст")
		return "", fmt.Errorf("повтор пуст")
	}

	validDate, err := time.Parse(internal.DateLayout, date)
	if err != nil {
		logger.LogMessage(fmt.Sprintf("[ERROR] Неправильная дата: %v", err))
		return "", fmt.Errorf("неправильная дата %v", err)
	}

	repeatParts := strings.Fields(repeat)
	if len(repeatParts) < 1 {
		logger.LogMessage("[ERROR] Неверное правило повторения")
		return "", fmt.Errorf("неверное правило повторения")
	}

	rule := repeatParts[0]

	var result string
	switch rule {
	case "d":
		if len(repeatParts) < 2 {
			logger.LogMessage("[ERROR] Отсутствует интервал для правила d")
			return "", fmt.Errorf("отсутствует интервал для правила d")
		}
		result, err = everyDay(now, validDate, repeatParts[1])
	case "y":
		result, err = everyYear(now, validDate)
	case "w":
		if len(repeatParts) < 2 {
			logger.LogMessage("[ERROR] Отсутствуют дни для правила w")
			return "", fmt.Errorf("отсутствуют дни для правила w")
		}
		result, err = everyWeek(validDate, now, repeatParts[1])
	case "m":
		if len(repeatParts) < 2 {
			logger.LogMessage("[ERROR] Отсутствуют дни для правила m")
			return "", fmt.Errorf("отсутствуют дни для правила m")
		}
		result, err = everyMonth(validDate, now, repeatParts[1:])
	default:
		logger.LogMessage(fmt.Sprintf("[ERROR] Неверное правило повторения: %v", rule))
		return "", fmt.Errorf("неверное правило повторения: %v", rule)
	}

	return result, err
}

func everyDay(now, date time.Time, daysStr string) (string, error) {
	d, err := strconv.Atoi(daysStr)
	if err != nil || d > 400 || d <= 0 {
		logger.LogMessage("[ERROR] Неверное правило повторения в d")
		return "", fmt.Errorf("неверное правило повторения в d")
	}

	resultDate := date.AddDate(0, 0, d)
	for resultDate.Before(now) {
		resultDate = resultDate.AddDate(0, 0, d)
	}

	return resultDate.Format(internal.DateLayout), nil
}

func everyWeek(date, now time.Time, daysStr string) (string, error) {
	days := strings.Split(daysStr, ",")
	validDays := make(map[int]bool)
	for _, day := range days {
		d, err := strconv.Atoi(day)
		if err != nil || d < 1 || d > 7 {
			logger.LogMessage(fmt.Sprintf("[ERROR] Неверный день недели: %s", day))
			return "", fmt.Errorf("неверный день недели: %s", day)
		}
		validDays[d] = true
	}

	date = date.AddDate(0, 0, 1)
	for {
		weekDay := int(date.Weekday())
		if weekDay == 0 {
			weekDay = 7
		}

		if validDays[weekDay] {
			return date.Format(internal.DateLayout), nil
		}
		date = date.AddDate(0, 0, 1)
	}
}

func everyMonth(date, now time.Time, days []string) (string, error) {
	month := date.Month()
	for {
		for _, dayStr := range days {
			targetDay, err := strconv.Atoi(dayStr)
			if err != nil || targetDay < 1 || targetDay > 31 {
				logger.LogMessage(fmt.Sprintf("[ERROR] Неверный день в правиле месяца: %v", dayStr))
				return "", fmt.Errorf("неверный день в правиле месяца: %v", dayStr)
			}

			newDate := time.Date(date.Year(), month, targetDay, 0, 0, 0, 0, time.Local)
			if newDate.Before(now) {
				continue
			}

			return newDate.Format(internal.DateLayout), nil
		}

		date = date.AddDate(0, 1, 0)
		month = date.Month()
	}
}

func everyYear(now, date time.Time) (string, error) {
	if date.Before(now) {
		for date.Before(now) {
			date = date.AddDate(1, 0, 0) // добавляем год
		}
	} else {
		date = date.AddDate(1, 0, 0) // добавляем год
	}

	return date.Format(internal.DateLayout), nil
}
