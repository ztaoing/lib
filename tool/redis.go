package tool

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

//连接redis
func RedisConnFactory(name string) (redis.Conn, error) {
	if ConfRedisMap != nil && ConfRedisMap.List != nil {
		for confName, cfg := range ConfRedisMap.List {
			//如果redis服务已经存在
			if name == confName {
				//在ProxyList随机获取一个host
				randHost := cfg.ProxyList[rand.Intn(len(cfg.ProxyList))]
				//未设置连接超时时间
				if cfg.ConnTimeout == 0 {
					cfg.ConnTimeout = 50
				}
				//未设置读超时时间
				if cfg.ReadTimeout == 0 {
					cfg.ReadTimeout = 100
				}
				//未设置写超时时间
				if cfg.WriteTimeout == 0 {
					cfg.WriteTimeout = 100
				}
				//使用以下指定的参数连接redis
				c, err := redis.Dial("tcp", randHost,
					redis.DialConnectTimeout(time.Duration(cfg.ConnTimeout)*time.Millisecond),
					redis.DialReadTimeout(time.Duration(cfg.ReadTimeout)*time.Millisecond),
					redis.DialWriteTimeout(time.Duration(cfg.WriteTimeout)*time.Millisecond))
				if err != nil {
					//建立失败
					return nil, err
				}
				//使用密码
				if cfg.Password != "" {
					if _, err := c.Do("AUTH", cfg.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				if cfg.Db != 0 {
					if _, err := c.Do("SELECT", cfg.Db); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil
			}

		}
	}
	return nil, errors.New("create redis conn failed")
}

//执行command并记录日志
func RedisLogDo(trace *TraceContext, c redis.Conn, commandName string, args ...interface{}) (interface{}, error) {
	startTime := time.Now()
	replay, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startTime).Seconds()),
		})
	} else {
		//将请求应答转换为string
		replyStr, _ := redis.String(replay, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startTime).Seconds()),
		})

	}
	return replay, err
}

//通过配置执行redis
func RedisConfDo(trace *TraceContext, name string, commandName string, args ...interface{}) (interface{}, error) {
	c, err := RedisConnFactory(name)
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method": commandName,
			"err":    errors.New("RedisConnFactory error:" + name),
		})
		return nil, err
	}
	defer c.Close()

	startTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startTime).Seconds()),
		})
	}
	return reply, nil
}
