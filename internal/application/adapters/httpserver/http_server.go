package httpserver

import (
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	engine *gin.Engine
}

func NewHTTPServer() *HTTPServer {
	server := &HTTPServer{
		engine: gin.New(),
	}

	server.Routes()
	return server
}

func (s *HTTPServer) Routes() {
	s.engine.POST("/submit")
	s.engine.GET("/list")
	s.engine.DELETE("/value/:id")
}

func (s *HTTPServer) Start() error {
	return s.engine.Run(":8080")
}
