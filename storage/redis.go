package storage

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

const (
	// OK is the standard response of a Redis server if everything went fine
	RedisOK = "OK"
)

type RedisStorage struct{}

type RedisPool struct {
	pool *redis.Pool
}

type RedisConnection struct {
	conn redis.Conn
}

func (rs *RedisStorage) NewPool(url, auth string) Pool {
	rp := RedisPool{
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", url)
				if err != nil {
					return nil, err
				}

				// If we don`t have an auth set, we don`t have to call redis
				if len(auth) == 0 {
					return c, err
				}

				if _, err := c.Do("AUTH", auth); err != nil {
					c.Close()
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}

	return rp
}

func (rp RedisPool) Close() error {
	return rp.pool.Close()
}

func (rp RedisPool) Get() Connection {
	rc := RedisConnection{
		conn: rp.pool.Get(),
	}
	return &rc
}

func (rc *RedisConnection) Close() error {
	return rc.conn.Close()
}

// MarkRepositoryAsTweeted marks a single projects as "already tweeted".
// This information will be stored in Redis as a simple set with a TTL.
// The timestamp of the tweet will be used as value.
func (rc *RedisConnection) MarkRepositoryAsTweeted(projectName, score string) (bool, error) {
	result, err := redis.String(rc.conn.Do("SET", projectName, score, "EX", GreyListTTL, "NX"))
	if result == RedisOK && err == nil {
		return true, err
	}
	return false, err
}

// IsRepositoryAlreadyTweeted checks if a project was already tweeted.
// If it is not available
//	a) the project was not tweeted yet
//	b) the project ttl expired and is ready to tweet again
func (rc *RedisConnection) IsRepositoryAlreadyTweeted(projectName string) (bool, error) {
	return redis.Bool(rc.conn.Do("EXISTS", projectName))
}
