package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError carries an HTTP status and a stable machine-readable code.
// Validation and business errors use HTTP 200 with success=false, matching
// the convention used by the cuberouter reference implementation.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewAPIError(code, message string) *APIError {
	return &APIError{Status: http.StatusOK, Code: code, Message: message}
}

func NewAPIErrorStatus(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

func Fail(c *gin.Context, err *APIError) {
	status := http.StatusOK
	if err != nil && err.Status != 0 {
		status = err.Status
	}
	code, message := "INTERNAL_ERROR", "Internal server error"
	if err != nil {
		code, message = err.Code, err.Message
	}
	c.JSON(status, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}

func AbortFail(c *gin.Context, err *APIError) {
	c.Abort()
	Fail(c, err)
}
