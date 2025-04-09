package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

type FSM struct {
	Db ports.DbConfig
}

func NewFSM(db ports.DbConfig) *FSM {
	return &FSM{Db: db}
}

func (fsm *FSM) Apply(logEntry *raft.Log) interface{} {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "node-Apply",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	kvpair := &models.KVPair{}

	// Unmarshal Data
	err := json.Unmarshal(logEntry.Data, kvpair)
	if err != nil {
		logger.Error("error unmarshaling data", "data", kvpair, "error", err)
		return fmt.Errorf("error Unmarshalling of log: %v", err)
	}

	// Sending Data To db
	// Added Retry mechanism for saving task to Db
	deadline := time.Now().Add(5 * time.Second)
	i := 0.0
	for time.Now().Before(deadline) {
		back_off := math.Pow(i, 3.0)*10 + 100
		err = fsm.Db.SaveToDb(kvpair)
		if err == nil {
			break
		}
		logger.Info("error Saving log to Db", "error", err)
		logger.Info("Retrying Apply to Db")
		time.Sleep(time.Duration(back_off))
		i++
	}

	logger.Info("Successfully written the value to Db", "value", kvpair)
	return nil
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (fsm *FSM) Restore(rc io.ReadCloser) error {
	return nil
}
