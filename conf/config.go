package conf

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
)

type AppConfig struct {
	MysqlConfig MysqlConfig `json:"mysql"`
	RedisConfig RedisConfig `json:"redis"`
	KafkaConfig KafkaConfig `json:"kafka"`
}

type MysqlConfig struct {
	Url string `json:"url"`
	User string	`json:"user"`
	Pass string	`json:"pass"`
}

type RedisConfig struct {
	Addr string `json:"addr"`
}

type KafkaConfig struct {
	Addr []string `json:addr`
	ProducerConfig ProducerConfig `json:"producer`
	ConsumerConfig ConsumerConfig `json:"consumer"`
}

type ProducerConfig struct {

}

type ConsumerConfig struct {
	GroupId string `json:"groupId"`
}

func init() {
	load("./conf/config.json")
}

var (
	appConfig AppConfig
)

func GetInstance() *AppConfig {
	return &appConfig
}

func load (fileName string) {
	data, err:= ioutil.ReadFile(fileName)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
	err = json.Unmarshal(data, &appConfig)
	if err != nil {
		logs.Error(err)
		panic(err)
	}
}