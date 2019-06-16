package redis

import (
	conf2 "envelop/conf"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
)

var (
	Client *redis.Client
)

func init () {

	conf:=conf2.GetInstance().RedisConfig

	opt := redis.Options{
		Addr:     conf.Addr,
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

