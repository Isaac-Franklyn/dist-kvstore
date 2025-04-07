package models

import (
	"github.com/hashicorp/raft"
)

type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
}

type KVPair struct {
	Key   string
	Value interface{}
}

type RaftNode struct {
	Node *raft.Raft
}
