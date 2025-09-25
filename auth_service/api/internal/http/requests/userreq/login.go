package userreq
//user_requests
import (
    "strings"
    "github.com/mymindmap/api/pkg/core/validator"
)

type Login struct {
    Email string `json:"email" validate:"required"`
    Password string `json:"password" validate:"required"`
}

func (r *Login) Validate() error {
    r.Email = strings.TrimSpace(r.Email)
    r.Password = strings.TrimSpace(r.Password)

    return validator.Validate(r)
}