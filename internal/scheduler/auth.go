package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go_final_project/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret    = []byte("my_jwt_secret_key")
	todoPassword string
)

func initConfig() {
	todoPassword = os.Getenv("TODO_PASSWORD")
}

func init() {
	initConfig()
}

type Credentials struct {
	Password string `json:"password"`
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.LogMessage("auth", fmt.Sprintf("[ERROR] Ошибка декодирования запроса: %v", err))
		http.Error(w, `{"error":"Неверный формат запроса"}`, http.StatusBadRequest)
		return
	}

	if todoPassword == "" || creds.Password != todoPassword {
		logger.LogMessage("auth", "[ERROR] Неверная попытка авторизации")
		http.Error(w, `{"error":"Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	passwordHash := sha256.Sum256([]byte(todoPassword))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"passwordHash": hex.EncodeToString(passwordHash[:]),
		"exp":          time.Now().Add(8 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		logger.LogMessage("auth", fmt.Sprintf("[ERROR] Ошибка создания JWT-токена: %v", err))
		http.Error(w, `{"error":"Ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	logger.LogMessage("auth", "[INFO] Пользователь успешно авторизован")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if todoPassword != "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				logger.LogMessage("auth", "[ERROR] Отсутствует токен в куки")
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				logger.LogMessage("auth", "[ERROR] Невалидный JWT-токен")
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.LogMessage("auth", "[ERROR] Ошибка получения данных из токена")
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			passwordHash := sha256.Sum256([]byte(todoPassword))
			if claims["passwordHash"] != hex.EncodeToString(passwordHash[:]) {
				logger.LogMessage("auth", "[ERROR] Токен не соответствует текущему паролю")
				http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
