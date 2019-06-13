package redis_test

import (
	"envelop/redis"
	"github.com/astaxie/beego/logs"
	"os"
	"testing"
	_ "envelop/redis"
	redis2 "github.com/go-redis/redis"
	"time"
)

func TestMain (m *testing.M) {

	os.Exit(m.Run())

}


func TestTxPipeline(t *testing.T) {

	redis.TxPipeline(func(pipeliner redis2.Pipeliner) error {
		res, err:= pipeliner.SetNX("hello","world", 100 * time.Second).Result()
		logs.Info("res", res, "err", err)
		return err
	})

}

