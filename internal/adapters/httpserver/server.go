package httpserver

import (
	"log"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	engine *gin.Engine
}

func NewHTTPServer(cluster ports.RaftClusterService) *HTTPServer {
	server := &HTTPServer{
		engine: gin.New(),
	}

	server.engine.Use(gin.Logger())
	server.engine.Use(gin.Recovery())
	server.SetupRoutes(cluster)
	return server
}

func (srv *HTTPServer) SetupRoutes(cluster ports.RaftClusterService) {
	srv.engine.POST("/submit", SubmitKV(cluster))
}

func (srv *HTTPServer) Start(PORT string) error {
	log.Printf("Starting server on port: %v...", PORT)
	return srv.engine.Run(PORT)
}
