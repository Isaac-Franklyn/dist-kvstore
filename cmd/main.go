package main

import (
	"os"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/adapters/database"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/adapters/httpserver"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/core/raftcluster"
	"github.com/hashicorp/go-hclog"
	"github.com/joho/godotenv"
)

func main() {

	// Setting up Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "main",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Error("error loading env files...", err)
		os.Exit(1)
	}

	// Environment variables for raft
	raftPeers := os.Getenv("RAFT_PEERS")
	httpPort := os.Getenv("PORT")
	raftAddr := os.Getenv("RAFT_CLUSTER_ADDRESS")
	raftPort := os.Getenv("RAFT_CLUSTER_PORT")
	dataDir := os.Getenv("DATA_DIR")

	// Environment variables for database
	dbUserName := os.Getenv("DB_USER_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	// Start the Db Server
	db, err := database.NewDbInstance(dbUserName, dbPassword, dbPort, dbName, dbHost)
	if err != nil {
		logger.Error("failed to connect Database", "error", err)
		os.Exit(1)
	}
	logger.Info("Successfuly started the Database Instance...")

	// Create the table
	if err := db.CreateTable(); err != nil {
		logger.Error("failed to create table", "error", err)
	}
	logger.Info("Successfully Created Database Table", "name", dbName)

	// Start the Raft Cluster
	raftCluster := raftcluster.NewRaftCluster()
	raftCluster.StartCluster(raftPeers, raftAddr, raftPort, dataDir, db)
	logger.Info("Successfully Started the Raft Cluster Nodes", "listening on...", "port", raftPort)

	// Start the HTTP Server
	srv := httpserver.NewHTTPServer(raftCluster)
	if err := srv.Start(httpPort); err != nil {
		logger.Error("failed to start the server", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully Started the Http Server...", "listening on", "port", httpPort)
}
