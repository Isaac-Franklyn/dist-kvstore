package database

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
)

func (db *DbInstance) CreateTable() error {

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "CreateTable",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	str := `
	CREATE TABLE dist_kvstore (
		id SERIAL PRIMARY KEY,
		key TEXT NOT NULL UNIQUE,
 		value VARCHAR(500) NOT NULL,
		created_at TIMESTAMP
	);`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.Db.ExecContext(ctx, str)
	if err != nil {
		logger.Error("failed to Create Table", "table query", str, "error", err)
		return err
	}

	logger.Info("Successfully created table", "table", "dist_kvstore")
	return nil
}
