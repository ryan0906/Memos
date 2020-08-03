package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ryan0906/Memos/pkg/protocol/grpc"
	v1 "github.com/ryan0906/Memos/pkg/service/v1"
)

// Config configuration for server
type Config struct {
	GRPCPort string

	DatastoreDBHost     string
	DatastoreDBUser     string
	DatastoreDBPassword string
	DatastoreDBSchema   string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "Database host")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "", "Database password")
	flag.StringVar(&cfg.DatastoreDBSchema, "db-schema", "", "Database schema")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: %s", cfg.GRPCPort)
	}

	param := "parseTime=True"

	dbSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DatastoreDBUser, cfg.DatastoreDBPassword, cfg.DatastoreDBHost, cfg.DatastoreDBSchema, param)

	db, err := sql.Open("sql", dbSource)
	if err != nil {
		return fmt.Errorf("Failed to open database, error: %v", err)
	}
	defer db.Close()

	v1API := v1.NewMemoServiceServer(db)
	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
