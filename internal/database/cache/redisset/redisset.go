package redisset

import "github.com/go-redis/redis"

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

func (r *RedisSet) Contains(name, value string) (bool, error) {
	return r.client.SIsMember(name, value).Result()
}

func (r *RedisSet) Clear(name string) error {
	return r.client.Del(name).Err()
}

func NewRedisSet(client *redis.Client) *RedisSet {
	return &RedisSet{
		client: client,
	}
}
