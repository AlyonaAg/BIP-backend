package keycache

import (
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis"
)

type Cache struct {
	config *Config
	cache  *redis.Client
}

func NewCache(config *Config) *Cache {
	return &Cache{
		config: config,
	}
}

func (c *Cache) Open() error {
	config, err := c.GetConfig()
	if err != nil {
		return err
	}

	log.Print(config.Port)
	cache := redis.NewClient(&redis.Options{
		Addr:     config.Port,
		Password: config.Password,
		DB:       0,
	})

	c.SetCache(cache)
	log.Print("Cache OK.")
	return nil
}

func (c *Cache) Set(key, value string) error {
	config, err := c.GetConfig()
	if err != nil {
		return err
	}

	cache, err := c.GetCache()
	if err != nil {
		return err
	}

	if err := cache.Set(key, value,
		time.Minute*time.Duration(config.ExpireDuration)).Err(); err != nil {
		return err
	}

	return nil
}

func (c *Cache) Get(key string) (string, error) {
	cache, err := c.GetCache()
	if err != nil {
		return "", err
	}

	val, err := cache.Get(key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *Cache) Del(key string) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	cache.Del(key)
	return nil
}

func (c *Cache) GetConfig() (*Config, error) {
	if c.config == nil {
		return nil, errors.New("empty cache config")
	}
	return c.config, nil
}

func (c *Cache) GetCache() (*redis.Client, error) {
	if c.cache == nil {
		return nil, errors.New("empty cache")
	}
	return c.cache, nil
}

func (c *Cache) SetCache(cache *redis.Client) {
	c.cache = cache
}
