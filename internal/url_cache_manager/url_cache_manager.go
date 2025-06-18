package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/limbo/url_shortener/internal/api"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	cli *redis.Client
}

type RedisConfig struct {
	Address  string
	Password string
}

func New(cfg RedisConfig) *Manager {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
	})
	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal(err)
	}
	return &Manager{
		cli: client,
	}
}

func (m *Manager) CacheLink(shortCode string, link string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := m.cli.Set(ctx, shortCode, link, time.Hour*6).Err()
	return err
}

func (m *Manager) GetLink(shortCode string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	value, err := m.cli.Get(ctx, shortCode).Result()
	if err == redis.Nil {
		return "", api.ErrNoKey
	} else if err != nil {
		return "", errors.New("error getting link: " + err.Error())
	}
	return value, nil
}
