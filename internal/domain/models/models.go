package models

import (
	"net"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

type KVPair struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type RaftNode struct {
	Node         *raft.Raft
	Id           string
	NodeAddr     string
	TcpAdvAddr   *net.TCPAddr
	TcpTransport raft.Transport
	SnapStore    raft.SnapshotStore
	LogStore     raft.LogStore
	StableStore  raft.StableStore
	Fsm          raft.FSM
	Logger       hclog.Logger
	IsJoined     bool
}
