package middleware

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"
	"errors"
    "io"
	"bytes"
	"github.com/mymindmap/api/internal/auth"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(authService *auth.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		 if authService == nil {
            http.Error(w, "Authentication service unavailable", http.StatusInternalServerError)
            return
        }
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Пробуем получить токен из cookie
			cookie, err := r.Cookie("auth_token")
			if err != nil || cookie.Value == "" {
				// Если нет токена, продолжаем без авторизации
				next.ServeHTTP(w, r)
				return
			}
			authHeader = "Bearer " + cookie.Value
		}

		// Проверяем формат заголовка
		if !strings.HasPrefix(authHeader, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		// Извлекаем токен
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Валидируем токен
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			// Если токен невалиден, продолжаем без авторизации
			next.ServeHTTP(w, r)
			return
		}

		// Добавляем пользователя в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
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

// RequireAuth middleware требует авторизации для доступа
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(UserContextKey)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequirePermission middleware проверяет права доступа
func RequirePermission(authService *auth.AuthService, object, action string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(UserContextKey)
			if claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userClaims := claims.(*auth.Claims)
			
			// Проверяем права доступа по роли пользователя
			if !authService.CheckPermission(userClaims.Role, object, action) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// GetAuthUser извлекает пользователя из контекста
func GetAuthUser(ctx context.Context) *auth.Claims {
	if user, ok := ctx.Value(UserContextKey).(*auth.Claims); ok {
		return user
	}
	return nil
}