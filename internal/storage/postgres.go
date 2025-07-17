package storage

import (
	"batch-saver/internal/models"
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repo struct {
	db *gorm.DB
}

func NewRepository(cfg Config) (*repo, error) {
	addr := cfg.address()
	db, err := gorm.Open(postgres.Open(addr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	m, err := migrate.New("file://internal/storage/migrations", addr)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	if err = m.Up(); err != nil {
		return nil, err
	}

	return &repo{db}, nil
}

func (repo *repo) Save(ctx context.Context, e []models.Event) error {
	return repo.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&e).
		Error
}

type Config struct {
	Host     string
	Port     int
	Db       string
	User     string
	Password string
	Ssl      bool
}

func (config Config) address() string {
	address := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=", config.User, config.Password, config.Host, config.Port, config.Db)

	if !config.Ssl {
		address += "disable"
	}

	return address
}
