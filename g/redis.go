package g

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

var RedisConnPool *redis.Pool

func InitRedisConnPool() {
	dsn := Config().Redis.Dsn
	maxIdle := Config().Redis.MaxIdle
	idleTimeout := 240 * time.Second

	connTimeout := time.Duration(Config().Redis.ConnTimeout) * time.Millisecond
	readTimeout := time.Duration(Config().Redis.ReadTimeout) * time.Millisecond
	writeTimeout := time.Duration(Config().Redis.WriteTimeout) * time.Millisecond

	RedisConnPool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", dsn, connTimeout, readTimeout, writeTimeout)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", Config().Redis.Passwd); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("[ERROR] ping redis fail", err)
	}
	return err
}
