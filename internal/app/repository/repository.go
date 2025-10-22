package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func New(dsn string, rdb *redis.Client) (*Repository, error) {
	logrus.Infof("Попытка подключения к PostgreSQL с DSN: %s", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	logrus.Info("Успешное подключение к PostgreSQL.")
	return &Repository{
		db:  db,
		rdb: rdb,
	}, nil
}

func (r *Repository) AddToBlacklist(ctx context.Context, token string, exp time.Duration) error {
	cmd := r.rdb.Set(ctx, token, "blacklist", exp)
	err := cmd.Err()
	if err != nil {
		logrus.Errorf("Не удалось добавить токен в блэклист. TTL: %v. Ошибка: %v", exp, err)
	}

	return err
}
