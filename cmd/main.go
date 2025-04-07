package main

import (
	"github.com/Isaac-Franklyn/dist-kvstore/internal/application/adapters/httpserver"
	raftcluster "github.com/Isaac-Franklyn/dist-kvstore/internal/application/raft"
)

func main() {

	cluster := raftcluster.NewRaftCluster()
	cluster.StartCluster(3)

	server := httpserver.NewHTTPServer(cluster)
	server.Start()

}
