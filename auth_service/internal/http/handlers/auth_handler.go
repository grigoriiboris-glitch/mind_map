package handlers

import (
	"log"
	"net/http"
	"strings"
	"log/slog"
	"encoding/base64"
    "encoding/json"
	"github.com/mymindmap/api/internal/auth"
	"github.com/mymindmap/api/internal/http/requests/userreq"
	"github.com/mymindmap/api/internal/http/middleware"
	"github.com/mymindmap/api/repositorybolt"
	resp "github.com/mymindmap/api/pkg/core/response"
)

type AuthHandler struct {
	authService *auth.AuthService
	userRepo    *repositorybolt.UserRepository
	logger      *log.Logger
}

func NewAuthHandler(authService *auth.AuthService, userRepo *repositorybolt.UserRepository, logger *log.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, userRepo: userRepo, logger: logger}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req userreq.Login
	if err := middleware.DecodeJSONBody(r, &req); err != nil {
	   resp.RespondErrors(w, http.StatusUnauthorized, err)
       return
    }

	tokenPair, err := h.authService.LoginUser(r.Context(), &req)
	if err != nil {
		resp.RespondErrors(w, http.StatusUnauthorized, err)
		return
	}

	config := h.authService.GetConfig()
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenPair.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Установите true для HTTPS
		MaxAge:   int(config.TokenExpiration.Seconds()),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Установите true для HTTPS
		MaxAge:   int(config.RefreshTokenExp.Seconds()),
	})


}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req userreq.Register
	if err := middleware.DecodeJSONBody(r, &req); err != nil {
       resp.RespondErrors(w, http.StatusUnauthorized, err)
       return
    }

	user, err := h.authService.RegisterUser(r.Context(), &req)
	if err != nil {
		resp.RespondErrors(w, http.StatusBadRequest, err)
		return
	}

	resp.RespondJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})

	resp.RespondJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie or header
	var refreshToken string
	
	// Try cookie first
	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	} else {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if refreshToken == "" {
		resp.RespondError(w, http.StatusUnauthorized, "refresh token is required")
		return
	}

	tokenPair, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		resp.RespondErrors(w, http.StatusUnauthorized, err)
		return
	}

	config := h.authService.GetConfig()
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenPair.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Установите true для HTTPS
		MaxAge:   int(config.TokenExpiration.Seconds()),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Установите true для HTTPS
		MaxAge:   int(config.RefreshTokenExp.Seconds()),
	})
	
	resp.RespondJSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"access_token": tokenPair.AccessToken,
		"expires_at":   tokenPair.ExpiresAt,
	})
}

func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetAuthUser(r.Context())
	if user == nil {
		resp.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	resp.RespondJSON(w, http.StatusOK, resp.Ok("ok"))
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetAuthUser(r.Context())
	if claims == nil {
	    slog.Info("auth service error:")
		resp.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.GetSecureUserByID(r.Context(), claims.UserID)

	if err != nil || user == nil {
		slog.Info("GetCurrentUser %s",err)
		resp.RespondError(w, http.StatusInternalServerError, "GetCurrentUser failed to get  user")
		return
	}
	user.Name = "Katy"
	w.Header().Set("X-User-Role", "user")
    jsonBytes, _ := json.Marshal(user)
    encoded := base64.StdEncoding.EncodeToString(jsonBytes)

    w.Header().Set("X-User-Body", encoded)
	resp.RespondJSON(w, http.StatusOK, user)
}
