package middleware

import (
	"go-boilerplate/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const requestIDHeader = "X-Request-Id"

// RequestIDMiddleware sets a request_id in the context and logger for each request
func RequestIDMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		logWithReqID := log.WithField("request_id", requestID)
		ctx = logger.WithLogger(ctx, logWithReqID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}
