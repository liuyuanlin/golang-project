package center

import (
	//"database/sql"
	//"github.com/golang/snappy"
	"fmt"

	"github.com/garyburd/redigo/redis"
	//"hash/crc32"
	//"io"
	"golang-project/dpsg/logger"
	//"golang-project/dpsg/proto"
	"strconv"
	//"stats"
	//"time"
	"golang-project/dpsg/common"
)

const (
	keylen = 64
)

func (self *Center) del(table string, key string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("DEL", table+":"+key)
	if err != nil {
		logger.Fatal("del error: %s (%s, %s, %d)", err.Error(), table, key)
	}

	return
}

func (self *Center) setInt(table string, key string, value int) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("SET", table+":"+key, value)
	if err != nil {
		logger.Fatal("setInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) getInt(table string, key string) (value int, err error) {
	if len(key) > keylen {
		return 0, fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := self.maincache.Get()
	defer cache.Recycle()

	value, err = redis.Int(cache.Conn.Do("GET", table+":"+key))
	if err != nil {
		if err != redis.ErrNil {
			logger.Fatal("getInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
		} else {
			err = nil
		}
	}

	return
}

func (self *Center) setString(table string, key string, value string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("SET", table+":"+key, value)
	if err != nil {
		logger.Fatal("setInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) getString(table string, key string) (value string, err error) {
	if len(key) > keylen {
		return "", fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := self.maincache.Get()
	defer cache.Recycle()

	value, err = redis.String(cache.Conn.Do("GET", table+":"+key))
	if err != nil {
		if err != redis.ErrNil {
			logger.Fatal("getInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
		} else {
			err = nil
		}
	}

	return
}

func (self *Center) setexpire(table string, key string, time string) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("EXPIRE", table+":"+key, time)
	if err != nil {
		logger.Fatal("PEXPIRE error: %s (%s, %s, %s)", err.Error(), table, key, time)
	}

	return
}

func (self *Center) sadd(table string, key string, value string) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("SADD", table+":"+key, value)
	if err != nil {
		logger.Fatal("sadd error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) srem(table string, key string, value string) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("SREM", table+":"+key, value)
	if err != nil {
		logger.Fatal("srem error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) exists(table string, key string) bool {
	cache := self.maincache.Get()
	defer cache.Recycle()

	exist, err := redis.Int(cache.Conn.Do("EXISTS", table+":"+key))
	if err != nil {
		logger.Fatal("exists error: %s (%s, %s, %d)", err.Error(), table, key, exist)
	}

	return exist == 1
}

func (self *Center) scard(table string, key string) (num int) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	num, err := redis.Int(cache.Conn.Do("SCARD", table+":"+key))
	if err != nil {
		logger.Fatal("scard error: %s (%s, %s, %d)", err.Error(), table, key, num)
	}

	return
}

func (self *Center) srandmember(table string, key string) (value string, err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	value, err = redis.String(cache.Conn.Do("SRANDMEMBER", table+":"+key))
	if err != nil {
		logger.Error("srandmember error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) zadd(table string, key string, value string, score uint32) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("ZADD", table+":"+key, strconv.FormatInt(int64(score), 10), value)
	if err != nil {
		logger.Fatal("zadd error: %s (%s, %s, %s, %d)", err.Error(), table, key, value, score)
	}

	return
}

func (self *Center) zrem(table string, key string, value string) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("ZREM", table+":"+key, value)
	if err != nil {
		logger.Fatal("zrem error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func (self *Center) zcard(table string, key string) (uint32, error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	length, err := redis.Int(cache.Conn.Do("zcard", table+":"+key))
	if err != nil && err != redis.ErrNil {
		logger.Fatal("zcard error: %s (%s, %s, %s, %d)", err.Error(), table, key, length)
	}

	return uint32(length), err
}

func (self *Center) zscore(table string, key string, value string) (uint32, error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	score, err := redis.Int(cache.Conn.Do("ZSCORE", table+":"+key, value))
	if err != nil && err != redis.ErrNil {
		logger.Fatal("zscore error: %s (%s, %s, %s, %d)", err.Error(), table, key, value, score)
	}

	return uint32(score), err
}

func (self *Center) zrevrange(table string, key string, start int, stop int) (rets []string, err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Conn.Do("ZREVRANGE", table+":"+key, start, stop))
	if err != nil {
		logger.Fatal("zrevrange error: %s (%s, %s, %d, %d)", err.Error(), table, key, start, stop)
	}

	return
}

func (self *Center) zrevrank(table string, key string, value string) (uint32, error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rank, err := redis.Int(cache.Conn.Do("ZREVRANK", table+":"+key, value))
	if err != nil && err != redis.ErrNil {
		logger.Fatal("zrevrank error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return uint32(rank), err
}

func (self *Center) keys(table string, key string) (rets []string, err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Conn.Do("keys", table+common.DbTableKeySplit+key))
	if err != nil {
		logger.Fatal("keys error: %s (%s, %s)", err.Error(), table, key)
	}

	return
}

func (self *Center) hset(table string, key string, field string, value string) (err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	_, err = cache.Conn.Do("hset", table+common.DbTableKeySplit+key, field, value)
	if err != nil {
		logger.Fatal("hset error: %s (%s, %s, %s, %s)", err.Error(), table, key, field, value)
	}

	return
}

func (self *Center) hgetall(table string, key string) (rets []string, err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Conn.Do("hgetall", table+common.DbTableKeySplit+key))
	if err != nil {
		logger.Fatal("keys error: %s (%s, %s)", err.Error(), table, key)
	}

	return
}

//add for seach myself
func (self *Center) zrank(table, key, value string) (rank int, err error) {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rank, err = redis.Int(cache.Conn.Do("ZRANK", table+":"+key, value))
	if err != nil {
		logger.Error("zrank error: %s (%s, %s)", err.Error(), table, value)
	}

	return
}
