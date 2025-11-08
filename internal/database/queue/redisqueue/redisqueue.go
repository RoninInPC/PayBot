package redisqueue

import (
	"errors"
	red "github.com/go-redis/redis"
	"main/internal/entity/mapper"
)

type RedisQueue[Anything any] struct {
	db        *red.Client
	queueName string
}

func (r RedisQueue[Anything]) RPush(value Anything) error {
	jsoned, err := mapper.ToJson(value)
	if err != nil {
		return errors.New("JSON mapper error " + err.Error())
	}
	return r.db.RPush(r.queueName, jsoned).Err()
}

func (r RedisQueue[Anything]) LPop() (*Anything, error) {
	bytes, err := r.db.LPop(r.queueName).Bytes()
	if err != nil {
		return nil, errors.New("LPOP error " + err.Error())
	}
	answer, err := mapper.FromJson[Anything](string(bytes))
	return &answer, err
}

func InitRedisQueue[Anything any](address, password, queueName string) *RedisQueue[Anything] {
	return &RedisQueue[Anything]{
		db: red.NewClient(&red.Options{
			Addr:     address,
			Password: password,
			DB:       0,
		}),
		queueName: queueName,
	}
}
