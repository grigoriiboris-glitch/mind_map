package response

import (
    "errors"
    "fmt"
    "strings"
   // "log"
    "net/http"
    "encoding/json"
    "github.com/go-playground/validator/v10"
)

type Response struct {
    Status bool        `json:"status"`
    Error  string      `json:"error,omitempty"`
    Data   interface{} `json:"data,omitempty"`
}

const (
    StatusOK    = true
    StatusError = false
)

func Error(msg string) Response {
    return Response{
        Status: StatusError,
        Error:  msg,
    }
}

func Ok(data interface{}) Response {
    return Response{
        Status: StatusOK,
        Data:   data,
    }
}

func Errors(e error) Response {
    var errMsgs []string

    // Пробуем найти ValidationErrors в цепочке ошибок
    var validationErrors validator.ValidationErrors
    if errors.As(e, &validationErrors) {
        for _, err := range validationErrors {
            switch err.ActualTag() {
            case "required":
                errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
            case "url":
                errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
            case "min":
                errMsgs = append(errMsgs, fmt.Sprintf("field %s must be at least %s characters", err.Field(), err.Param()))
            case "max":
                errMsgs = append(errMsgs, fmt.Sprintf("field %s must be at most %s characters", err.Field(), err.Param()))
            case "email":
                errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid email", err.Field()))
            default:
                errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid: %s", err.Field(), err.ActualTag()))
            }
        }
    } else {
        // Если это не ошибки валидации, просто возвращаем текст ошибки
        errMsgs = append(errMsgs, e.Error())
    }

    return Response{
        Status: false,
        Error:  strings.Join(errMsgs, ", "),
    }
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		//logger.Printf("json encode error: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, msg string) {
    RespondJSON(w, status, Response{
        Status: false,
        Error:  msg,
    })
}

func RespondErrors(w http.ResponseWriter, status int, e error) {
    RespondJSON(w, status, Errors(e))
}