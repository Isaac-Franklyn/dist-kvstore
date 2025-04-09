package raftcluster

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

type RaftCluster struct {
	Cluster []*models.RaftNode
}

func NewRaftCluster() *RaftCluster {
	return &RaftCluster{}
}

func (c *RaftCluster) StartCluster(RAFT_PEERS, RAFT_ADDR, RAFT_PORT, DATA_DIR string, db ports.DbConfig) error {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "Cluster Service",
		Level: hclog.Debug,
	})

	raftPeers, err := strconv.Atoi(RAFT_PEERS)
	if err != nil {
		logger.Error("error converting RAFT_PEERS to int, RAFT_PEERS: %v", RAFT_PEERS)
		os.Exit(1)
	}
	raftPort, err := strconv.Atoi(RAFT_PORT)
	if err != nil {
		logger.Error("error converting RAFT_PEERS to int, RAFT_PORT: %v", RAFT_PORT)
		os.Exit(1)
	}

	if err := os.MkdirAll(DATA_DIR, 0700); err != nil {
		logger.Error("Failed to create data directory", "path", DATA_DIR, "error", err)
		os.Exit(1)
	}

	for i := 0; i < raftPeers; i++ {
		addr := fmt.Sprintf("%v:%v", RAFT_ADDR, raftPort+i)
		id := fmt.Sprintf("node-%v", i+1)
		node, err := CreateRaftNode(addr, id, DATA_DIR, i, db)
		if err != nil {
			logger.Error("error creating node", "id", id, "error", err)
			os.Exit(1)
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

	logger.Info("Waiting for leader election....")
	time.Sleep(time.Second * 1)

	logger.Info("Leader elected", "leaderAddr", c.Cluster[0].Node.Leader())

	for i := 1; i < raftPeers; i++ {
		addFuture := c.Cluster[0].Node.AddVoter(
			raft.ServerID(c.Cluster[i].Id),
			raft.ServerAddress(c.Cluster[i].NodeAddr),
			0,
			0,
		)
		if err := addFuture.Error(); err != nil {
			logger.Error("error adding node to the server", "node", c.Cluster[i].Id, "addr", c.Cluster[i].NodeAddr)
			os.Exit(1)
		}
		c.Cluster[i].IsJoined = true
		logger.Info("successfully added node to the server", "node", c.Cluster[i].Id, "addr", c.Cluster[i].NodeAddr)
	}

	logger.Info("successfuly bootstraped the first node...")

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

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "SendValueToCluster",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

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
		logger.Error("timeout getting leader", "error", err, "node", node)
		return err
	}

	node.Logger.Info("Current leader", node.Id, "taking in requests from", node.NodeAddr)

	// Marshal the data
	byteValue, err := json.Marshal(val)
	if err != nil {
		logger.Error("failed to marshal val to bytes", "value", val, "error", err)
	}

	// Apply the byte value to the leader node
	deadline := time.Now().Add(5 * time.Second)
	i := 0.0
	for time.Now().Before(deadline) {
		applyFuture := node.Node.Apply(byteValue, time.Second*2)
		if err := applyFuture.Error(); err != nil {
			back_off := math.Pow(i, 3.0)*10 + 100
			logger.Info("error applying to leader", "leader", node.Id, "value", val, "error", err)
			logger.Info("Retrying Apply to Leader...", "leader", node.Id, "Value", val, "retry", i)
			time.Sleep(time.Millisecond * time.Duration(back_off))
			i++
			continue
		}
		logger.Info("Successfully applied data to the cluster", "leader", node.Id, "value", val)
		return nil
	}

	return fmt.Errorf("apply to leader %s failed after retries", node.Id)
}
