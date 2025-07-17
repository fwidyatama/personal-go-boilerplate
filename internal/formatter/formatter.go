package formatter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code      int         `json:"code"`
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

func Respond(c *gin.Context, status int, success bool, message string, data interface{}, errMsg string, metadata interface{}) {
	resp := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     errMsg,
		Code:      status,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	c.JSON(status, resp)
}

func Success(c *gin.Context, data interface{}, message string) {
	Respond(c, http.StatusOK, true, message, data, "", nil)
}

func Error(c *gin.Context, status int, message, errMsg string) {
	Respond(c, status, false, message, nil, errMsg, nil)
}
