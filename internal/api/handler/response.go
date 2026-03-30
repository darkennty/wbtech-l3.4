package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

type resultResponse struct {
	Status int            `json:"status"`
	Result map[string]any `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func ReturnErrorResponse(ctx *gin.Context, logger zlog.Zerolog, statusCode int, message string) {
	logger.Error().Msg(message)
	ctx.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func ReturnResultResponse(ctx *gin.Context, status int, result map[string]any) {
	ctx.JSON(http.StatusOK, resultResponse{
		Status: status,
		Result: result,
	})
}
