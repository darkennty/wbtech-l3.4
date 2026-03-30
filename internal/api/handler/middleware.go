package handler

import (
	"fmt"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func handlerFunc(logger zlog.Zerolog, f ginext.HandlerFunc) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		log := fmt.Sprintf("%s %s %s", c.Request.Method, c.Request.RequestURI, time.Now().Format(time.RFC3339))
		logger.Info().Msg(log)
		f(c)
	}
}
