package redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gpmgo/gopm/modules/log"
	"time"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = ""
)

// newRedisPool:创建 Redis 链接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,                //最大链接数量
		MaxActive:   30,                //最大活跃连接数量
		IdleTimeout: 300 * time.Second, //超时关闭时间
		Dial: func() (conn redis.Conn, e error) {
			//1.打开连接
			c, e := redis.Dial("tcp", redisHost)
			if e != nil {
				log.Error(e.Error())
				return nil, e
			}
			//2.访问认证  //如果没有设置密码不用执行这句话
			//if _, e = c.Do("AUTH", redisPass); e != nil {
			//	c.Close()
			//	return nil, e
			//}
			return c, e
		},
		// 每分钟去检测一下 redis 链接状况.如果出错了 redis 会自动关闭 shutdown
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//初始化
func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
