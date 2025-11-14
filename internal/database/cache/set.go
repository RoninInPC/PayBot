package cache

import (
	"github.com/go-redis/redis"
)

type Set interface {
	Add(string) error
	GetAll() ([]string, error)
	Delete(string) error
}

type RedisSet struct {
	client *redis.Client
	key    string
}

func NewRedisSet(client *redis.Client, key string) *RedisSet {
	return &RedisSet{
		client: client,
		key:    key,
	}
}

func (r *RedisSet) Add(userID string) error {
	// SADD - добавляем пользователя в множество
	return r.client.SAdd(r.key, userID).Err()
}

func (r *RedisSet) GetAll() ([]string, error) {
	// SMEMBERS - получаем всех пользователей
	return r.client.SMembers(r.key).Result()
}

func (r *RedisSet) Delete(userID string) error {
	// SREM - удаляем пользователя из множества
	return r.client.SRem(r.key, userID).Err()
}
