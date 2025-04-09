package ports

import models "github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"

type RaftClusterService interface {
	SendValueToCluster(value *models.KVPair) error
}

type RaftConfigService interface {
	StartCluster()
	CreateNode()
}
