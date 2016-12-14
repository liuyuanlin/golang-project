package accountserver

import (
	"database/sql"
	"errors"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/stats"
	"hash/crc32"
	"io"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	keylen = 64
)

type table struct {
	name         string
	caches       cacheGroup
	dbs          dbGroup
	deleteExpiry uint64
	tableStats   *stats.Timings
	qpsRates     *stats.Rates
}

func NewTable(name string, cfg common.TableConfig, db *AccountServer) (t *table) {
	var caches cacheGroup
	var dbs dbGroup

	if cfg.CacheProfile != "" {
		var exist bool
		if caches, exist = db.cacheGroups[cfg.CacheProfile]; !exist {
			logger.Fatal("NewTable: table cache profile not found: %s", cfg.CacheProfile)
		}
	}

	if cfg.DBProfile != "" {
		var exist bool
		if dbs, exist = db.dbGroups[cfg.DBProfile]; !exist {
			logger.Fatal("NewTable: table db profile not found: %s", cfg.DBProfile)
		}

		for _, dbpool := range dbs {
			db := dbpool.Get()
			defer db.Recycle()

			query := fmt.Sprintf(`
					CREATE TABLE IF NOT EXISTS %s (
				    id BINARY(64) NOT NULL PRIMARY KEY,
				    relateid BINARY(64),
				    KEY (relateid)
				) ENGINE=InnoDB;
				`, name)

			logger.Info("CreateQuery :%s", query)
			rst, err := db.Exec(
				query,
			)

			if err != nil {
				logger.Fatal("NewTable: db %v create table %s faild! %s", dbpool, name, err.Error())
			}

			logger.Info("NewTable: db %v init %s: %v", dbpool, name, rst)

		}
	}

	if caches == nil && dbs == nil {
		logger.Fatal("NewTable: table %s need a save func", name)
	}

	queryStats := stats.NewTimings("")
	qpsRates := stats.NewRates("", queryStats, 20, 10e9)
	return &table{
		name, caches, dbs,
		cfg.DeleteExpiry,
		queryStats,
		qpsRates,
	}
}

func (self *table) write(key string, value string) (err error) {
	if len(key) > keylen || len(value) > keylen {
		return fmt.Errorf("key (%s) (%s)len must <= 64", key, value)
	}

	defer self.tableStats.Record("write", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := hid % uint32(len(self.caches))
		cache := self.caches[cidx].Get()
		defer cache.Recycle()

		numbers, err := redis.Int(cache.Conn.Do("HSETNX", self.name, key, value))
		if err != nil {
			logger.Fatal("write error: %s (%s, %v)", err.Error(), key, value)
		}
		if numbers == 0 {
			return errors.New("HSETNX failed")
		}

		//反向
		_, err = cache.Conn.Do("HSET", self.name+"_re", value, key)
		if err != nil {
			logger.Fatal("write re error: %s (%s, %v)", err.Error(), key, value)
		}
	}

	if self.dbs != nil {
		didx := hid % uint32(len(self.dbs))
		db := self.dbs[didx].Get()
		defer db.Recycle()

		_, err = db.Exec("INSERT INTO "+self.name+" (id, relateid) values(?, ?) ON DUPLICATE KEY UPDATE relateid=?;", key, value, value)
		logger.Info("mysql exec success!")
		if err != nil {
			logger.Fatal("write error: %s (%s, %v)", err.Error(), key, value)
		}
	}

	return
}

func (self *table) get(key string) (ret string, err error) {
	if len(key) > keylen {
		return "", fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("get", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := hid % uint32(len(self.caches))
		cache := self.caches[cidx].Get()
		defer cache.Recycle()

		ret, err = redis.String(cache.Conn.Do("HGET", self.name, key))
		//fmt.Print("GET ", self.name+"_"+key, "\n")
		if err != nil {
			if err != redis.ErrNil {
				logger.Fatal("get error: %s (%s, %v)", err.Error(), key, ret)
			} else {
				err = nil
			}
		}

		if ret != "" {
			return
		}
	}

	if self.dbs != nil {
		didx := hid % uint32(len(self.dbs))
		db := self.dbs[didx].Get()
		defer db.Recycle()

		var rows *sql.Rows
		rows, err = db.Query("SELECT relateid from "+self.name+" where id = CAST(? as BINARY(64)) LIMIT 1;", key)

		if err != nil {
			logger.Fatal("get error: %s (%s, %v)", err.Error(), key, rows)
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&ret)
			if err != nil {
				logger.Fatal("get scan error %s (%s)", err.Error(), key)
			}
			//去掉前后的空格
			logger.Info("before trim : %s, %d", ret, len(ret))
			ret = strings.TrimRight(ret, string(byte(0)))
			logger.Info("end trim : %s, %d", ret, len(ret))
			return
		}
	}

	return "", nil
}

func (self *table) del(key string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("del", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := hid % uint32(len(self.caches))
		cache := self.caches[cidx].Get()
		defer cache.Recycle()

		//反向查询
		svalue, err := redis.String(cache.Conn.Do("HGET", self.name, key))
		if err != nil {
			if err != redis.ErrNil {
				logger.Fatal("reget error: %s (%s)", err.Error(), key)
			} else {
				err = nil
				svalue = ""
			}
		}

		_, err = cache.Conn.Do("HDEL", self.name, key)
		if err != nil {
			if err != redis.ErrNil {
				logger.Fatal("del error: %s (%s)", err.Error(), key)
			} else {
				err = nil
			}
		}

		//反向删除
		if svalue != "" {
			_, err = cache.Conn.Do("HDEL", self.name+"_re", svalue)
			if err != nil {
				if err != redis.ErrNil {
					logger.Fatal("redel error: %s (%s)", err.Error(), key)
				} else {
					err = nil
				}
			}
		}
	}

	if self.dbs != nil {
		didx := hid % uint32(len(self.dbs))
		db := self.dbs[didx].Get()
		defer db.Recycle()

		_, err = db.Query("DELETE from "+self.name+" where id = CAST(? as BINARY(64));", key)

		if err != nil {
			logger.Fatal("delete error: %s (%s)", err.Error(), key)
		}
	}

	return
}

//反向查询
func (self *table) reget(key string) (ret string, err error) {
	if len(key) > keylen {
		return "", fmt.Errorf("key (%s) len must <= 64", key)
	}

	defer self.tableStats.Record("get", time.Now())
	hid := makeHash(key)

	if self.caches != nil {
		cidx := hid % uint32(len(self.caches))
		cache := self.caches[cidx].Get()
		defer cache.Recycle()

		ret, err = redis.String(cache.Conn.Do("HGET", self.name+"_re", key))
		//fmt.Print("GET ", self.name+"_"+key, "\n")
		if err != nil {
			if err != redis.ErrNil {
				logger.Fatal("get error: %s (%s, %v)", err.Error(), key, ret)
			} else {
				err = nil
			}
		}

		if ret != "" {
			return
		}
	}

	if self.dbs != nil {
		didx := hid % uint32(len(self.dbs))
		db := self.dbs[didx].Get()
		defer db.Recycle()

		var rows *sql.Rows
		rows, err = db.Query("SELECT id from "+self.name+" where relateid = CAST(? as BINARY(64)) LIMIT 1;", key)

		if err != nil {
			logger.Fatal("get error: %s (%s, %v)", err.Error(), key, rows)
		}

		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&ret)
			if err != nil {
				logger.Fatal("get scan error %s (%s)", err.Error(), key)
			}
			//去掉前后的空格
			ret = strings.TrimRight(ret, string(byte(0)))
			return
		}
	}

	return "", nil
}

func makeHash(key string) uint32 {
	ieee := crc32.NewIEEE()
	io.WriteString(ieee, key)
	return ieee.Sum32()
}
