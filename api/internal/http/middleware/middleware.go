package middleware

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"
	"errors"
    "io"
	"log/slog"
	"bytes"
	"encoding/base64"
	"github.com/mymindmap/api/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// вытаскивает данные из заголовков
func AuthHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
claims := r.Header.Get("X-User-Body")
var cl models.User

// Декодируем base64
decoded, err := base64.StdEncoding.DecodeString(claims)
if err != nil {
    http.Error(w, "invalid User response", http.StatusBadRequest)
    return
}
slog.Info("Decoded user data:", string(decoded))

// Распаковываем JSON
err = json.Unmarshal(decoded, &cl)
if err != nil {
    http.Error(w, "invalid user data", http.StatusBadRequest)
    return
}

        // Добавляем пользователя в контекст
        ctx := context.WithValue(r.Context(), UserContextKey, cl)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func JSONContentType(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
            contentType := r.Header.Get("Content-Type")
            if !strings.HasPrefix(contentType, "application/json") {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnsupportedMediaType)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "content-type must be application/json",
                })
                return
            }

            // Декодируем тело в map для базовой проверки JSON
            var body map[string]interface{}
            
            // Сохраняем оригинальное тело для повторного чтения
            bodyBytes, err := io.ReadAll(r.Body)
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "failed to read request body",
                })
                return
            }
            
            // Восстанавливаем тело для дальнейшего использования
            r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
            
            // Проверяем JSON на валидность
            if err := json.Unmarshal(bodyBytes, &body); err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "invalid json format",
                })
                return
            }

            // Сохраняем декодированные данные в контекст
            ctx := context.WithValue(r.Context(), "jsonBody", body)
            ctx = context.WithValue(ctx, "jsonBodyBytes", bodyBytes)
            r = r.WithContext(ctx)
        }
        next.ServeHTTP(w, r)
    })
}

// Дополнительные функции для работы с декодированными данными:

// GetJSONBody извлекает декодированный JSON из контекста
func GetJSONBody(r *http.Request) (map[string]interface{}, bool) {
    body, ok := r.Context().Value("jsonBody").(map[string]interface{})
    return body, ok
}

// GetJSONBodyBytes извлекает сырые байты JSON из контекста
func GetJSONBodyBytes(r *http.Request) ([]byte, bool) {
    bodyBytes, ok := r.Context().Value("jsonBodyBytes").([]byte)
    return bodyBytes, ok
}

// DecodeJSONBody декодирует JSON в конкретную структуру (удобно для хендлеров)
func DecodeJSONBody(r *http.Request, v interface{}) error {
    bodyBytes, ok := GetJSONBodyBytes(r)
    if !ok {
        return errors.New("json body not available")
    }
    return json.Unmarshal(bodyBytes, v)
}

// GetAuthUser извлекает пользователя из контекста
func GetAuthUser(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}