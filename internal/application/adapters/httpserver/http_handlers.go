package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/models"
	"github.com/Isaac-Franklyn/dist-kvstore/internal/domain/ports"
	"github.com/gin-gonic/gin"
)

func SubmitValue(cluster ports.RaftService) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		cmd := &models.Command{}
		body := ctx.Request.Body
		decoder := json.NewDecoder(body)
		if err := decoder.Decode(cmd); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := cluster.SendValueToCluster(cmd); err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"cmd": cmd})
	}
}
