package httpserver

import (
	"encoding/json"
	"net/http"
	"os"

	models "github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
)

func SubmitKV(cluster ports.RaftClusterService) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		logger := hclog.New(&hclog.LoggerOptions{
			Name:   "KvService",
			Level:  hclog.Debug,
			Output: os.Stdout,
		})

		kvPair := &models.KVPair{}

		decoder := json.NewDecoder(ctx.Request.Body)
		if err := decoder.Decode(kvPair); err != nil {
			logger.Error("error decoding object", "object", kvPair, "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := cluster.SendValueToCluster(kvPair)
		if err != nil {
			logger.Error("error sending value to cluster", "error", err)
		}

		logger.Info("Successfully submitted value to cluster...", "value", kvPair)
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": kvPair})
	}

}
