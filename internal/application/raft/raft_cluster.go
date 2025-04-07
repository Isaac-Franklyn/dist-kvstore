package raftcluster

import (
	"encoding/json"
	"fmt"
	"log"

	"time"

	core "github.com/Isaac-Franklyn/dist-kvstore/internal/core/fsm"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type RaftCluster struct {
	Cluster []*models.RaftNode
}

func NewRaftCluster() *RaftCluster {
	return &RaftCluster{}
}

func (cluster *RaftCluster) StartCluster(n int) error {

	for i := 0; i < n; i++ {
		addr := fmt.Sprint("127.0.0.1:", 9000+i)
		id := fmt.Sprintf("node%v", i+1)
		node, err := CreateNode(id, addr)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		cluster.Cluster = append(cluster.Cluster, node)
	}

	config := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID("node1"),
				Address: raft.ServerAddress("127.0.0.1:9001"),
			},
		},
	}
	future := cluster.Cluster[0].Node.BootstrapCluster(config)
	if err := future.Error(); err != nil {
		log.Fatal("error bootstraping to cluster")
	}

	// Add Node 2
	if err := cluster.Cluster[0].Node.AddVoter("node2", "127.0.0.1:9002", 0, 0).Error(); err != nil {
		log.Fatalf("Failed to add node2: %v", err)
	}

	// Add Node 3
	if err := cluster.Cluster[0].Node.AddVoter("node3", "127.0.0.1:9003", 0, 0).Error(); err != nil {
		log.Fatalf("Failed to add node3: %v", err)
	}

	return nil
}

func CreateNode(id, addr string) (*models.RaftNode, error) {
	// 1. Config
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)

	// 2. Transport
	transport, err := raft.NewTCPTransport(addr, nil, 3, raft.DefaultConfig().CommitTimeout, nil)
	if err != nil {
		return nil, fmt.Errorf("tcp transport error: %v", err)
	}

	// 3. BoltDB store for logs and stable store
	logStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("raft-log-%s.db", id))
	if err != nil {
		return nil, fmt.Errorf("log store error: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("raft-stable-%s.db", id))
	if err != nil {
		return nil, fmt.Errorf("stable store error: %v", err)
	}

	// 4. Snapshot store
	snapshots, err := raft.NewFileSnapshotStore(".", 2, nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot store error: %v", err)
	}

	// 5. FSM (use your FSM struct)
	fsm := core.NewFSM() // you'll create this in internal/core/fsm.go

	// 6. Create the Raft instance
	node, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("raft creation error: %v", err)
	}

	raftNode := &models.RaftNode{
		Node: node,
	}

	return raftNode, nil
}

func (cluster *RaftCluster) SendValueToCluster(cmd *models.Command) error {
	errchan := make(chan error, 1)
	nodechan := make(chan *models.RaftNode, 1)

	i := 0
	go func() {
		for {
			node, err := cluster.GetLeader()
			if err != nil {
				expBackoff := (1 << i) * 100
				time.Sleep(time.Millisecond * time.Duration(expBackoff))
				i++
				continue
			}
			nodechan <- node
			errchan <- nil
			break
		}
	}()

	var node *models.RaftNode
	select {
	case node = <-nodechan:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for leader")
	}

	go func(cmd *models.Command, node *models.RaftNode) {
		bytes, err := json.Marshal(cmd)
		if err != nil {
			errchan <- fmt.Errorf("failed to marshal: %v", err)
			return
		}

		future := node.Node.Apply(bytes, 5*time.Second)
		if err := future.Error(); err != nil {
			errchan <- fmt.Errorf("apply error: %v", err)
			return
		}

		errchan <- nil
	}(cmd, node)

	return <-errchan
}

func (cluster *RaftCluster) GetLeader() (*models.RaftNode, error) {
	for _, node := range cluster.Cluster {
		if node.Node.State() == raft.Leader {
			return node, nil
		}
	}

	return &models.RaftNode{}, fmt.Errorf("no leader is available")
}
