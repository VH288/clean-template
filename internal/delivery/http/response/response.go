package response

import "github.com/gin-gonic/gin"

type Body struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Send(c *gin.Context, code int, message string, data any) {
	c.JSON(code, Body{
		Message: message,
		Data:    data,
	})
}
