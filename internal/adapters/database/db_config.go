package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
)

type DbInstance struct {
	Db *sql.DB
}

func NewDbInstance(username, password, port, dbname, dbhost string) (*DbInstance, error) {

	connStr := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=disable", username, password, dbhost, port, dbname)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "Db-instance",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	d, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("failed to open connection to postgres", "conn-string", connStr, "error", err)
		return &DbInstance{}, fmt.Errorf("failed to open connection to postgres")
	}

	logger.Info("Successfully connected to Datbase", "db-name", dbname)
	dbinst := &DbInstance{
		Db: d,
	}

	return dbinst, nil
}
