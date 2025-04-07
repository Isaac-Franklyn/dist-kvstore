package ports

import "github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"

type RaftService interface {
	SendValueToCluster(value *models.KVPair) error
	GetLeader()
}
