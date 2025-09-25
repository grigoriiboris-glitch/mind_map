package handlers

import (
    //"net/http"
    "github.com/mymindmap/api/internal/auth"
    //"github.com/mymindmap/api/internal/http/middleware"
    "github.com/mymindmap/api/repository"
)

type UserHandler struct {
    userRepo    *repository.UserRepo
    authService *auth.AuthService
}

func NewUserHandler(userRepo *repository.UserRepo, authService *auth.AuthService) *UserHandler {
    return &UserHandler{userRepo: userRepo, authService: authService}
}