package config

import (
	"batch-saver/internal/service"
	"batch-saver/internal/storage"
)

type Config struct {
	LogLevel            string
	GRPCServerPort      int
	MaxConcurrentWrites int
	ServiceCfg          service.Config
	PostgresCfg         storage.Config
}
