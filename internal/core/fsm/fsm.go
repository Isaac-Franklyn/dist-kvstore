package fsm

import (
	"io"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/hashicorp/raft"
)

type FSM struct {
	Db ports.DbConfig
}

func NewFSM(db ports.DbConfig) *FSM {
	return &FSM{Db: db}
}

func (fsm *FSM) Apply(logEntry *raft.Log) interface{} {
	return nil
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (fsm *FSM) Restore(rc io.ReadCloser) error {
	return nil
}
