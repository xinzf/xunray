package storage

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Redis *_redis

func init() {
	Redis = new(_redis)
}

type _redis struct {
	client *redis.Client
}

func (this *_redis) Init() {
	addr := viper.GetString("redis.addr")
	if addr == "" {
		return
	}

	this.client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
	})

	_, err := this.client.Ping().Result()
	if err != nil {
		logrus.Fatalln("Redis connecte failed.")
	} else {
		logrus.Debugln("Redis inited.")
	}
}

func (this *_redis) Client() *redis.Client {
	return this.client
}
