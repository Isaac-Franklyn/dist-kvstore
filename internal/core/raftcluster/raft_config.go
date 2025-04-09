package raftcluster

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/core/fsm"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func CreateRaftNode(addr, id, Datadir string, i int, db ports.DbConfig) (*models.RaftNode, error) {

	randomnumber := 100 * (i + 1)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:       id,
		Level:      hclog.Debug,
		Output:     os.Stdout,
		JSONFormat: false,
	})

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)
	config.Logger = logger
	config.CommitTimeout = time.Millisecond * 100
	config.ElectionTimeout = time.Second * 5
	config.HeartbeatTimeout = time.Millisecond * time.Duration(randomnumber)
	config.SnapshotInterval = time.Second * 10
	config.SnapshotThreshold = 2

	// Creating a TCP Transport
	advAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Error("Failed to Resolve TCP Address", "error", err)
		return nil, fmt.Errorf("failed to resolve tcp address %v", err)
	}
	transport, err := raft.NewTCPTransport(addr, advAddr, 3, time.Second*10, os.Stderr)
	if err != nil {
		logger.Error("failed to create TCP Transport network to", id, "error", err)
		return nil, fmt.Errorf("failed to create tcp transport network %v", err)
	}

	// Creating a Stable Store
	stabledir := Datadir + "/raft-stablestore"
	stablepath := fmt.Sprintf("%v/%v.bolt", stabledir, id)
	if err := os.MkdirAll(stablepath, 0700); err != nil {
		logger.Error("failed to create directory", "directory", stablepath, "error", err)
		return nil, fmt.Errorf("failed to create parent directory: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(stablepath)
	if err != nil {
		logger.Error("error creating stable store with file path: ", stablepath)
		return nil, fmt.Errorf("failed to create stable store %v", err)
	}

	// Creating a Log Store
	logdir := Datadir + "/raft-logstore"
	logpath := fmt.Sprintf("%v/%v.bolt", logdir, id)
	if err := os.MkdirAll(logpath, 0700); err != nil {
		logger.Error("failed to create directory", logpath, "error", err)
		return nil, fmt.Errorf("failed to create parent directory: %v", err)
	}

	logstore, err := raftboltdb.NewBoltStore(logpath)
	if err != nil {
		logger.Error("error creating log-store with filepath: ", logpath)
		return nil, fmt.Errorf("failed to create log-store %v", err)
	}

	// Creating snapshot dir
	snapdir := Datadir + "/raft-snapstore"
	snappath := fmt.Sprintf("%v/%v", snapdir, id)
	if err := os.MkdirAll(snappath, 0700); err != nil {
		logger.Error("failed to create snapshot directory", "path", snappath, "error", err)
		return nil, fmt.Errorf("failed to create snapshot directory: %v", err)
	}

	// Create the snapshot store
	snapstore, err := raft.NewFileSnapshotStore(snappath, 2, os.Stderr)
	if err != nil {
		logger.Error("error creating snapstore with file path: ", snappath)
		return nil, fmt.Errorf("failed to create snapstore %v", err)
	}

	fsm := fsm.NewFSM(db)

	node, err := raft.NewRaft(config, fsm, logstore, stableStore, snapstore, transport)
	if err != nil {
		logger.Error("failed to created node: ", id)
		os.Exit(1)
	}

	raftNode := &models.RaftNode{
		Node:         node,
		Id:           id,
		NodeAddr:     addr,
		TcpAdvAddr:   advAddr,
		TcpTransport: transport,
		SnapStore:    snapstore,
		StableStore:  stableStore,
		LogStore:     logstore,
		Fsm:          fsm,
		Logger:       logger,
		IsJoined:     false,
	}

	return raftNode, nil

}
