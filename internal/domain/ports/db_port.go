package ports

import "github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"

type DbConfig interface {
	CreateTable() error
	SaveToDb(val *models.KVPair) error
}
