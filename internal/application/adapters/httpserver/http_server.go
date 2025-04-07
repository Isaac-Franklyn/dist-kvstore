package httpserver

import (
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	engine  *gin.Engine
	cluster ports.RaftService
}

func NewHTTPServer(cluster ports.RaftService) *HTTPServer {
	server := &HTTPServer{
		engine:  gin.New(),
		cluster: cluster,
	}

	server.Routes()
	return server
}

func (s *HTTPServer) Routes() {
	s.engine.POST("/submit", SubmitValue(s.cluster))
	s.engine.GET("/list")
	s.engine.DELETE("/value/:id")
}

func (s *HTTPServer) Start() error {
	return s.engine.Run(":8080")
}
