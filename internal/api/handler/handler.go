package handler

import (
	"WBTech_L3.4/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type Handler struct {
	services *service.Service
	logger   zlog.Zerolog
}

func NewHandler(services *service.Service, logger zlog.Zerolog) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitRoutes() *ginext.Engine {
	r := ginext.New("")

	r.POST("/upload", handlerFunc(h.logger, h.handleUpload))
	r.GET("/image/:id", handlerFunc(h.logger, h.handleGetImage))
	r.DELETE("/image/:id", handlerFunc(h.logger, h.handleDeleteImage))

	r.Static("/static", "./web")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	return r
}
