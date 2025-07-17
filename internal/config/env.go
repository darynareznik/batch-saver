package config

import (
	"batch-saver/internal/service"
	"batch-saver/internal/storage"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func Read() Config {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("APP")
	viper.SetDefault("LOG_LEVEL", "INFO")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_USERNAME", "postgres")
	viper.SetDefault("POSTGRES_DB_NAME", "events")
	viper.SetDefault("POSTGRES_SSL", false)
	viper.SetDefault("GRPC_PORT", 3000)
	viper.SetDefault("MAX_CONCURRENT_WRITES", 5)
	viper.SetDefault("BATCH_MAX_SIZE", 3)
	viper.SetDefault("BATCH_FLUSH_TIMEOUT", "1s")

	return Config{
		LogLevel:            viper.GetString("LOG_LEVEL"),
		GRPCServerPort:      viper.GetInt("GRPC_PORT"),
		MaxConcurrentWrites: viper.GetInt("MAX_CONCURRENT_WRITES"),
		ServiceCfg: service.Config{
			BatchMaxSize:      viper.GetInt("BATCH_MAX_SIZE"),
			BatchFlushTimeout: viper.GetDuration("BATCH_FLUSH_TIMEOUT"),
		},
		PostgresCfg: storage.Config{
			Host:     viper.GetString("POSTGRES_HOST"),
			Port:     viper.GetInt("POSTGRES_PORT"),
			Db:       viper.GetString("POSTGRES_DB_NAME"),
			User:     viper.GetString("POSTGRES_USERNAME"),
			Password: viper.GetString("POSTGRES_PWD"),
			Ssl:      viper.GetBool("POSTGRES_SSL"),
		},
	}
}

func (c Config) GetLogLevel() zerolog.Level {
	l, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		return zerolog.InfoLevel
	}

	return l
}
