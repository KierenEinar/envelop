package redis

import (
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
)

var (
	Client *redis.Client
)

func init () {
	opt := redis.Options{
		Addr:     "localhost:6379",
		PoolSize: 100,
		DB: 0,
	}
	Client = redis.NewClient(&opt)
	pong, err := Client.Ping().Result()
	logs.Info("redis conenct ping", pong, "err", err)
}

func TxPipeline (fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	pipe := Client.TxPipeline()
	err:= fn(pipe)
	if err !=nil {
		return nil, err
	}
	cmd, err:= pipe.Exec()
	return cmd, err
}

