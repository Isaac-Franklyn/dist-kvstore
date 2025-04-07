package ports

import "github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"

type RaftService interface {
	SendValueToCluster(cmd *models.Command) error
	GetLeader() (*models.RaftNode, error)
}
