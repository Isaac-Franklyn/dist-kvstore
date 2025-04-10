package raftcluster

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

type RaftCluster struct {
	Cluster []*models.RaftNode
	logger  hclog.Logger
}

func NewRaftCluster() *RaftCluster {
	return &RaftCluster{
		logger: hclog.New(&hclog.LoggerOptions{
			Name:   "Cluster Service",
			Level:  hclog.Debug,
			Output: os.Stdout,
		}),
	}
}

func (c *RaftCluster) StartCluster(RAFT_PEERS, RAFT_ADDR, RAFT_PORT, DATA_DIR string,
	db ports.DbConfig,
	nodeHBtimeout_inms,
	nodeElectimeout_insecs,
	nodeSnapshotInterval_insecs,
	nodeSnapshotthreshold,
	nodeCommittimeout_inms string) error {

	// Converting environment variables to strings.
	nodeHBTO_inms, err := strconv.Atoi(nodeHBtimeout_inms)
	if err != nil {
		c.logger.Error("error converting nodeheartbeat timout to int", "error", err)
		return fmt.Errorf("error converting nodeheartbeat timeout to int: %w", err)
	}
	nodeElecTO_insecs, err := strconv.Atoi(nodeElectimeout_insecs)
	if err != nil {
		c.logger.Error("error converting node election timout to int", "error", err)
		return fmt.Errorf("error converting nodeelection timeout to int: %w", err)
	}
	nodeSSInterval_insecs, err := strconv.Atoi(nodeSnapshotInterval_insecs)
	if err != nil {
		c.logger.Error("error converting nodeSnapshot interval timout to int", "error", err)
		return fmt.Errorf("error converting nodeSnapshot interval timeout to int: %w", err)
	}
	nodeSSthreshold, err := strconv.Atoi(nodeSnapshotthreshold)
	if err != nil {
		c.logger.Error("error converting nodesnapshot threshold to int", "error", err)
		return fmt.Errorf("error converting nodesnapshhot threshold to int: %w", err)
	}
	nodeCommTO_inms, err := strconv.Atoi(nodeCommittimeout_inms)
	if err != nil {
		c.logger.Error("error converting nodecommit timeout to int", "error", err)
		return fmt.Errorf("error converting nodecommit timeout to int: %w", err)
	}

	// Converting the number of peers from string to int
	raftPeers, err := strconv.Atoi(RAFT_PEERS)
	if err != nil {
		c.logger.Error("error converting RAFT_PEERS to int, RAFT_PEERS: %v", RAFT_PEERS)
		return fmt.Errorf("error converting RAFT_PEERS to int: %w", err)
	}

	// had to be converted to int because of the fact that it will be different for different nodes
	raftPort, err := strconv.Atoi(RAFT_PORT)
	if err != nil {
		c.logger.Error("error converting RAFT_PORT to int, RAFT_PORT: %v", RAFT_PORT)
		return fmt.Errorf("error converting RAFT_PORT to int: %w", err)
	}

	// Creating raft-data dir for storing logstores, stablestores, snapshot store
	if err := os.MkdirAll(DATA_DIR, 0700); err != nil {
		c.logger.Error("Failed to create data directory", "path", DATA_DIR, "error", err)
		return fmt.Errorf("error failing to create data directory: %w", err)
	}

	// Creating directory for LogStores
	logstorepath := filepath.Join(DATA_DIR, "/logstore")
	if err := os.MkdirAll(logstorepath, 0700); err != nil {
		c.logger.Error("Failed to create log store directory", "path", logstorepath, "error", err)
		return fmt.Errorf("error failing to create log store directory: %w", err)
	}

	// Creating directory for Stable Stores
	stablestorepath := filepath.Join(DATA_DIR, "/stablestore")
	if err := os.MkdirAll(stablestorepath, 0700); err != nil {
		c.logger.Error("Failed to create stable store directory", "path", stablestorepath, "error", err)
		return fmt.Errorf("error failing to create stable store directory: %w", err)
	}

	// Creating directorty for Snapshots
	snapshotstorepath := filepath.Join(DATA_DIR, "/snapshotstore")
	if err := os.MkdirAll(snapshotstorepath, 0700); err != nil {
		c.logger.Error("Failed to create snapshot store directory", "path", snapshotstorepath, "error", err)
		return fmt.Errorf("error failing to create snapshot store directory: %w", err)
	}

	for i := 0; i < raftPeers; i++ {
		addr := fmt.Sprintf("%v:%v", RAFT_ADDR, raftPort+i)
		id := fmt.Sprintf("node-%v", i+1)
		node, err := CreateRaftNode(addr,
			id,
			logstorepath,
			stablestorepath,
			snapshotstorepath,
			i,
			db,
			nodeHBTO_inms,
			nodeElecTO_insecs,
			nodeSSInterval_insecs,
			nodeSSthreshold,
			nodeCommTO_inms)
		if err != nil {
			c.logger.Error("error creating node", "id", id, "error", err)
			return fmt.Errorf("error creating node of id: %v, error: %w", id, err)
		}
		c.Cluster = append(c.Cluster, node)
	}

	raftConfig := raft.Configuration{
		Servers: []raft.Server{
			{ID: raft.ServerID(c.Cluster[0].Id), Address: raft.ServerAddress(c.Cluster[0].NodeAddr)},
		},
	}
	bootFuture := c.Cluster[0].Node.BootstrapCluster(raftConfig)
	if err := bootFuture.Error(); err != nil {
		c.Cluster[0].Logger.Error("failed to bootstrap0", "node", c.Cluster[0].Id, "addr", c.Cluster[0].NodeAddr)
		return fmt.Errorf("failed to bootstrap: %v", err)
	}
	c.Cluster[0].IsJoined = true

	c.logger.Info("Waiting for leader election....")
	time.Sleep(time.Second * 1)

	c.logger.Info("Leader elected", "leaderAddr", c.Cluster[0].Node.Leader())

	for i := 1; i < raftPeers; i++ {
		addFuture := c.Cluster[0].Node.AddVoter(
			raft.ServerID(c.Cluster[i].Id),
			raft.ServerAddress(c.Cluster[i].NodeAddr),
			0,
			0,
		)

		if err := addFuture.Error(); err != nil {
			c.logger.Error("error adding node to the server", "node", c.Cluster[i].Id, "addr", c.Cluster[i].NodeAddr)
			return fmt.Errorf("error adding node to the server: %w", err)
		}
		c.Cluster[i].IsJoined = true
		c.logger.Info("successfully added node to the server", "node", c.Cluster[i].Id, "addr", c.Cluster[i].NodeAddr)
	}

	c.logger.Info("successfuly bootstraped the first node...")

	return nil
}

func (c *RaftCluster) GetLeader() (*models.RaftNode, error) {

	nodechan := make(chan *models.RaftNode, 1)
	errchan := make(chan error, 1)

	go func() {
		for _, node := range c.Cluster {
			if node.Node.State() == raft.Leader {
				nodechan <- node
				errchan <- nil
				return
			}
		}
		nodechan <- nil
		errchan <- fmt.Errorf("failed to Get Leader")
	}()

	return <-nodechan, <-errchan
}

func (c *RaftCluster) SendValueToCluster(val *models.KVPair) error {

	leaderchan := make(chan *models.RaftNode, 1)
	errchan := make(chan error, 1)

	go func() {
		duration := time.Now().Add(5 * time.Second)
		i := 0.0
		for time.Now().Before(duration) {
			exp_backoff := math.Pow(i, 3.0)*10 + 100
			leader, err := c.GetLeader()
			leaderchan <- leader
			if leader == nil || err != nil {
				time.Sleep(time.Millisecond * time.Duration(exp_backoff))
				errchan <- err
				i++
				continue
			}
		}
		errchan <- fmt.Errorf("failed to get leader, timeout")
	}()

	node := <-leaderchan
	err := <-errchan
	if err != nil || node == nil {
		c.logger.Error("timeout getting leader", "error", err, "node", node)
		return err
	}

	node.Logger.Info("Current leader", node.Id, "taking in requests from", node.NodeAddr)

	// Marshal the data
	byteValue, err := json.Marshal(val)
	if err != nil {
		c.logger.Error("failed to marshal val to bytes", "value", val, "error", err)
	}

	// Apply the byte value to the leader node
	deadline := time.Now().Add(5 * time.Second)
	i := 0.0
	for time.Now().Before(deadline) {
		applyFuture := node.Node.Apply(byteValue, time.Second*2)
		if err := applyFuture.Error(); err != nil {
			back_off := math.Pow(i, 3.0)*10 + 100
			c.logger.Info("error applying to leader", "leader", node.Id, "value", val, "error", err)
			c.logger.Info("Retrying Apply to Leader...", "leader", node.Id, "Value", val, "retry", i)
			time.Sleep(time.Millisecond * time.Duration(back_off))
			i++
			continue
		}
		c.logger.Info("Successfully applied data to the cluster", "leader", node.Id, "value", val)
		return nil
	}

	return fmt.Errorf("apply to leader %s failed after retries", node.Id)
}
