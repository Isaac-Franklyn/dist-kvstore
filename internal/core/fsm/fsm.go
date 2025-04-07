package core

import (
	"encoding/json"
	"io"
	"log"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"

	"github.com/hashicorp/raft"
)

type FSM struct {
	Store map[string]interface{}
}

func NewFSM() *FSM {
	return &FSM{
		Store: make(map[string]interface{}),
	}
}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {

	var cmd models.Command

	if err := json.Unmarshal(logEntry.Data, &cmd); err != nil {
		log.Printf("Failed to unmarshal command: %v", err)
		return nil
	}

	switch cmd.Op {
	case "SET":
		f.Store[cmd.Key] = cmd.Value
	case "DEL":
		delete(f.Store, cmd.Key)
	}
	return "successfully applied value to the store"
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {

	return &fsmSnapshot{}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	return nil
}

type fsmSnapshot struct{}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (s *fsmSnapshot) Release() {}
