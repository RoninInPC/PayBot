package cache

import (
	"github.com/go-redis/redis"
)

type Sets interface {
	Add(name, value string) error
	GetAll(name string) ([]string, error)
	Delete(name, value string) error
	Clear(name string) error
}

type RedisSet struct {
	client *redis.Client
}

func (r *RedisSet) GetAll(name string) ([]string, error) {
	return r.client.SMembers(name).Result()
}

func (r *RedisSet) Add(name, value string) error {
	return r.client.SAdd(name, value).Err()
}

func (r *RedisSet) Delete(name, value string) error {
	return r.client.SRem(name, value).Err()
}

func (r *RedisSet) Clear(name string) error {
	return r.client.Del(name).Err()
}

func NewRedisSet(client *redis.Client) *RedisSet {
	return &RedisSet{
		client: client,
	}
}
