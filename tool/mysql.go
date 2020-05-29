package tool

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/e421083458/gorm"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

//初始化连接池
func InitDBPool(path string) error {
	//普通的db方式
	DbConfMap := &MysqlMapConf{}
	err := ParseConfig(path, DbConfMap)
	if err != nil {
		return err
	}
	if len(DbConfMap.List) == 0 {
		fmt.Printf("[INFO]%s%s\n", time.Now().Format(TimeFomat), "  mysql config is empty")
	}
	DBMapPool = map[string]*sql.DB{}
	GORMMapPool = map[string]*gorm.DB{}

	//通过获取到的配置设置连接池的信息
	for confName, DbConf := range DbConfMap.List {
		dbPool, err := sql.Open("mysql", DbConf.DataSourceName)
		if err != nil {
			return err
		}
		//设置最 数
		dbPool.SetMaxOpenConns(DbConf.MaxOpenConn)
		dbPool.SetMaxIdleConns(DbConf.MaxIdleConn)
		dbPool.SetConnMaxLifetime(time.Duration(DbConf.MaxConnLifeTime) * time.Second)
		//探测MySQL是否可以连接
		err = dbPool.Ping()
		if err != nil {
			return err
		}

		//gorm连接方式
		dbGorm, err := gorm.Open("mysql", DbConf.DataSourceName)
		if err != nil {
			return err
		}
		//默认使用单个表
		dbGorm.SingularTable(true)
		//当打印日志的时候获得上下文
		dbGorm.LogCtx(true)
		dbGorm.SetLogger(&MysqlGormLogger{Trace: NewTrace()})
		dbGorm.DB().SetMaxIdleConns(DbConf.MaxIdleConn)
		dbGorm.DB().SetMaxOpenConns(DbConf.MaxOpenConn)
		dbGorm.DB().SetConnMaxLifetime(time.Duration(DbConf.MaxConnLifeTime) * time.Second)
		DBMapPool[confName] = dbPool
		GORMMapPool[confName] = dbGorm

	}
	//手动配置连接
	if dbpool, err := GetDBPool("default"); err == nil {
		DBDefaultPool = dbpool
	}

	if dbpool, err := GetGormPool("default"); err == nil {
		GORMDefaultPool = dbpool
	}
	return nil
}

func GetDBPool(name string) (*sql.DB, error) {
	if dbpool, ok := DBMapPool[name]; ok {
		return dbpool, nil
	}
	return nil, errors.New("GetDBPool error")
}

func GetGormPool(name string) (*gorm.DB, error) {
	if dbpool, ok := GORMMapPool[name]; ok {
		return dbpool, nil
	}
	return nil, errors.New("GetGormPool error")
}

func CloseDB() error {
	for _, dbpool := range DBMapPool {
		dbpool.Close()
	}
	for _, dbpool := range GORMMapPool {
		dbpool.Close()
	}
	return nil
}

//执行query
func DBPoolLogQuery(trace *TraceContext, sqlDB *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	startExecTime := time.Now()
	rows, err := sqlDB.Query(query, args...)
	endExecTime := time.Now()
	if err != nil {
		//出错
		Log.TagError(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		//query成功
		Log.TagInfo(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return rows, err
}

//MySQL日志打印类
type MysqlGormLogger struct {
	gorm.Logger
	Trace *TraceContext
}

func (logger *MysqlGormLogger) NowFunc() time.Time {
	return time.Now()
}

//对数据进行格式化
func (logger *MysqlGormLogger) LogFormatter(values ...interface{}) (messages map[string]interface{}) {
	if len(values) > 1 {
		var (
			sql             string
			formattedValues []string
			level           = values[0]
			currentTime     = logger.NowFunc().Format("2006-01-02 15:04:05")
			source          = fmt.Sprintf("%v", values[1])
		)
		messages = map[string]interface{}{
			"level":        level,
			"source":       source,
			"current_time": currentTime,
		}
		//TODO
		if level == "sql" {
			messages["proc_time"] = fmt.Sprintf("%fs", values[2].(time.Duration).Seconds())

			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					//如果是时间
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); logger.isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}

			}
			if regexp.MustCompile(`\$\d+`).MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				for index, value := range regexp.MustCompile(`\?`).Split(values[3].(string), -1) {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}
			messages["sql"] = sql
			if len(values) > 5 {
				messages["affected_row"] = strconv.FormatInt(values[5].(int64), 10)
			}

		} else {
			messages["ext"] = values
		}
	}
	return
}

func (logger *MysqlGormLogger) isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func (logger *MysqlGormLogger) Print(values ...interface{}) {
	message := logger.LogFormatter(values...)
	if message["level"] == "sql" {
		Log.TagInfo(logger.Trace, "_com_mysql_success", message)
	} else {
		Log.TagInfo(logger.Trace, "_com_mysql_failure", message)
	}
}

//logCtx(true)会执行该方法
func (logger *MysqlGormLogger) CtxPrint(s *gorm.DB, values ...interface{}) {
	ctx, ok := s.GetCtx()
	trace := newTrace()
	if ok {
		trace = ctx.(*TraceContext)
	}
	message := logger.LogFormatter(values...)
	if message["level"] == "sql" {
		Log.TagInfo(trace, "_com_mysql_success", message)
	} else {
		Log.TagInfo(trace, "_com_mysql_failure", message)
	}

}
