package response

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool
	Message string
	Error   interface{}
	Data    interface{}
}

func SuccessResponse(ctx *gin.Context, statusCode int, message string, data ...interface{}) {

	response := Response{
		Status:  true,
		Message: message,
		Error:   nil,
		Data:    data,
	}
	ctx.JSON(statusCode, response)
}

func ErrorResponse(ctx *gin.Context, statusCode int, message string, err error, data interface{}) {

	errFields := strings.Split(err.Error(), "\n")
	response := Response{
		Status:  false,
		Message: message,
		Error:   errFields,
		Data:    data,
	}

	ctx.JSON(statusCode, response)
}
