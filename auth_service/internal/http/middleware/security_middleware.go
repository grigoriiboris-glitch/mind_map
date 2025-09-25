package middleware

import (
    "encoding/json"
    "net/http"
    "strings"
)

// SecurityConfig конфигурация для security middleware
type SecurityConfig struct {
    ExpectedHosts    []string // Разрешенные хосты (например, ["api.example.com", "localhost:8080"])
    EnableCSP        bool     // Включить Content Security Policy
    EnableHSTS       bool     // Включить HSTS (только для HTTPS)
    EnableFrameGuard bool     // Запретить embedding в iframe
}

// DefaultSecurityConfig дефолтная конфигурация
var DefaultSecurityConfig = SecurityConfig{
    ExpectedHosts:    []string{"localhost", "127.0.0.1"},
    EnableCSP:        true,
    EnableHSTS:       false, // Включать только на продакшене с HTTPS
    EnableFrameGuard: true,
}

// SecurityMiddleware добавляет security headers и проверяет host
func SecurityMiddleware(config SecurityConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Проверка Host header
            if len(config.ExpectedHosts) > 0 && !isValidHost(r.Host, config.ExpectedHosts) {
                respondSecurityError(w, "Invalid host header")
                return
            }

            // Устанавливаем security headers
            setSecurityHeaders(w, config)

            next.ServeHTTP(w, r)
        })
    }
}

// isValidHost проверяет что Host header соответствует ожидаемым значениям
func isValidHost(host string, expectedHosts []string) bool {
    // Убираем порт если есть
    hostWithoutPort := strings.Split(host, ":")[0]
    
    for _, expected := range expectedHosts {
        expectedWithoutPort := strings.Split(expected, ":")[0]
        
        if hostWithoutPort == expectedWithoutPort || 
           host == expected || 
           expected == "*" {
            return true
        }
    }
    return false
}

// setSecurityHeaders устанавливает security headers
func setSecurityHeaders(w http.ResponseWriter, config SecurityConfig) {
    headers := w.Header()

    // Запрет embedding в iframe
    if config.EnableFrameGuard {
        headers.Set("X-Frame-Options", "DENY")
    }

    // Content Security Policy
    if config.EnableCSP {
        csp := "default-src 'self'; " +
            "connect-src *; " +
            "font-src *; " +
            "script-src-elem * 'unsafe-inline'; " +
            "img-src * data:; " +
            "style-src * 'unsafe-inline'; " +
            "frame-ancestors 'none';"
        headers.Set("Content-Security-Policy", csp)
    }

    // XSS Protection
    headers.Set("X-XSS-Protection", "1; mode=block")

    // HSTS (только для HTTPS!)
    if config.EnableHSTS {
        headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    }

    // Referrer Policy
    headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

    // No sniffing
    headers.Set("X-Content-Type-Options", "nosniff")

    // Permissions Policy
    headers.Set("Permissions-Policy", 
        "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), "+
        "magnetometer=(), gyroscope=(), fullscreen=(self), payment=()")
}

// respondSecurityError отправляет ошибку безопасности
func respondSecurityError(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    
    errorResponse := map[string]string{
        "error":   "security_violation",
        "message": message,
    }
    
    json.NewEncoder(w).Encode(errorResponse)
}