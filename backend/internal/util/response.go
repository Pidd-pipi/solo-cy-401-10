// Package util provides shared helpers: unified responses and JWT.
package util

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/repository"
)

var validate = validator.New()

// Response is the unified API envelope.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK writes a successful response.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeSuccess, Message: constants.MsgOK, Data: data})
}

// Fail maps an error to a unified JSON error response.
func Fail(c *gin.Context, err error) {
	var appErr *constants.AppError
	if errors.As(err, &appErr) {
		c.JSON(http.StatusOK, Response{Code: appErr.Code, Message: appErr.Message, Data: nil})
		return
	}

	status := http.StatusInternalServerError
	code := constants.CodeInternal
	message := constants.MsgInternalError

	switch {
	case errors.Is(err, repository.ErrNotFound):
		status, code, message = http.StatusOK, constants.CodeNotFound, constants.MsgNotFound
	case errors.Is(err, constants.ErrUnauthorized):
		status, code, message = http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized
	case errors.Is(err, constants.ErrForbidden):
		status, code, message = http.StatusForbidden, constants.CodeForbidden, constants.MsgForbidden
	case errors.Is(err, constants.ErrConflict):
		status, code, message = http.StatusConflict, constants.CodeConflict, err.Error()
	case errors.Is(err, constants.ErrInvalidInput):
		status, code, message = http.StatusBadRequest, constants.CodeBadRequest, err.Error()
	}

	c.JSON(status, Response{Code: code, Message: message, Data: nil})
}

// BindAndValidate parses JSON and runs validator/v10 rules.
func BindAndValidate(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Fail(c, constants.NewAppError(constants.CodeBadRequest, constants.MsgInvalidParams))
		return false
	}
	if err := validate.Struct(obj); err != nil {
		Fail(c, constants.NewAppError(constants.CodeBadRequest, constants.MsgValidationFail+": "+firstValidationMessage(err)))
		return false
	}
	return true
}

func firstValidationMessage(err error) string {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) && len(verr) > 0 {
		return verr[0].Field() + " " + verr[0].Tag()
	}
	return ""
}
