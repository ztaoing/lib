package tool

import (
	"bytes"
	"database/sql"
	"github.com/e421083458/gorm"
	"github.com/spf13/viper"
	"io/ioutil"
	dlog "lib/log"
	"os"
	"strings"
	"time"
)

type BaseConf struct {
	DebugMode    string    `mapstructure:"debug_mode"`
	TimeLocation string    `mapstructrue:"time_location"` //时区
	Log          LogConfig `mapstructure:"log"`
	Base         struct {
		DebugMode    string `mapstructure:"debug_mode"`
		TimeLocation string `mapstructure:"time_location"`
	} `mapstructure:"base"`
}

type LogConfConsoleWriter struct {
	On    bool `mapstructure:"on"`
	Color bool `mapstructure:"color"`
}
type LogConfFileWriter struct {
	On              bool   `mapstructure:"on"`
	LogPath         string `mapstructure:"log_path"`
	RotateLogPath   string `mapstructure:"rotate_log_path"`
	WfLogPath       string `mapstructure:"wf_log_path"`
	RotateWfLogPath string `mapstructure:"rotate_log_path"`
}

type LogConfig struct {
	Level string               `mapstructure:"log_level"`
	FW    LogConfFileWriter    `mapstructure:"file_writer"`
	CW    LogConfConsoleWriter `mapstructure:"console_writer"`
}

//msyql

type MysqlMapConf struct {
	List map[string]*MYSQLConf `mapstructure:"list"`
}

type MYSQLConf struct {
	DriverName      string `mapstructure:"driver_name"`        //驱动名称
	DataSourceName  string `mapstructure:"data_source_name"`   //数据资源名
	MaxOpenConn     int    `mapstructure:"max_open_conn"`      //最大连接数
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`      //最大空闲连接
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"` //连接最大生存时间

}

//redis

type RedisMapConf struct {
	List map[string]*RedisConf `mapstructure:"list"`
}

type RedisConf struct {
	ProxyList    []string `mapstructure:"proxy_list"`
	Password     string   `mapstructure:"password"`
	Db           int      `mapstructure:"db"`
	ConnTimeout  int      `mapstructure:"conn_timeout"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

//全局变量
var ConfBase *BaseConf
var DBMapPool map[string]*sql.DB
var GORMMapPool map[string]*gorm.DB // 此gorm封装了上下文
var DBDefaultPool *sql.DB
var GORMDefaultPool *gorm.DB
var ConfRedis *RedisConf
var ConfRedisMap *RedisMapConf
var ViperConfMap map[string]*viper.Viper

//获取基本配置信息
func GetBaseConf() *BaseConf {
	return ConfBase
}

func InitBaseConf(path string) error {
	ConfBase = &BaseConf{}
	err := ParseConfig(path, ConfBase)
	if err != nil {
		return err
	}
	//debug模式
	if ConfBase.DebugMode == "" {
		if ConfBase.Base.DebugMode != "" {
			ConfBase.DebugMode = ConfBase.Base.DebugMode
		} else {
			ConfBase.DebugMode = "debug"
		}
	}

	if ConfBase.TimeLocation == "" {
		if ConfBase.Base.TimeLocation != "" {
			ConfBase.TimeLocation = ConfBase.Base.TimeLocation
		} else {
			ConfBase.TimeLocation = "Asia/Chongqing"
		}
	}
	if ConfBase.Log.Level == "" {
		ConfBase.Log.Level = "trace"
	}

	//配置日志
	logConf := dlog.LogConfig{
		Level: ConfBase.Log.Level,
		FW: dlog.ConfFileWriter{
			On:              ConfBase.Log.FW.On,
			LogPath:         ConfBase.Log.FW.LogPath,
			RotateLogPath:   ConfBase.Log.FW.RotateLogPath,
			WfLogPath:       ConfBase.Log.FW.WfLogPath,
			RotateWfLogPath: ConfBase.Log.FW.RotateWfLogPath,
		},
		CW: dlog.ConfConsoleWriter{
			On:    ConfBase.Log.CW.On,
			Color: ConfBase.Log.CW.Color,
		},
	}

	//使用配置设置log
	if err := dlog.SetupDefaultWithConf(logConf); err != nil {
		panic(err)
	}
	dlog.SetLayout("2006-01-02T15:04:05.000")
	return nil

}

func InitRedisConf(path string) error {
	ConfRedis := &RedisMapConf{}
	err := ParseConfig(path, ConfRedis)
	if err != nil {
		return err
	}
	ConfRedisMap = ConfRedis
	return nil
}

//初始化配置文件
func InitViperConf() error {
	f, err := os.Open(ConfEnvPath + "/")
	if err != nil {
		return err
	}
	fileList, err := f.Readdir(1024)
	if err != nil {
		return err
	}
	//将读取到的配置设置到对应类型的ViperConfMap中，例如：mysql_map :viper
	for _, f0 := range fileList {
		if !f0.IsDir() {
			bts, err := ioutil.ReadFile(ConfEnvPath + "/" + f0.Name())
			if err != nil {
				return err
			}
			v := viper.New()
			v.SetConfigType("toml")
			v.ReadConfig(bytes.NewBuffer(bts))
			pathArr := strings.Split(f0.Name(), ".")
			if ViperConfMap == nil {
				ViperConfMap = make(map[string]*viper.Viper)
			}
			//mysql_map :viper
			ViperConfMap[pathArr[0]] = v
		}
	}
	return nil
}

//获取配置信息
func GetConf(key string) interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.Get(strings.Join(keys[1:len(keys)], "."))
	return conf
}

func GetBoolConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	BoolConf := v.GetBool(strings.Join(keys[1:len(keys)], "."))
	return BoolConf
}

func GetFloat64Conf(key string) float64 {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	Float64Conf := v.GetFloat64(strings.Join(keys[1:len(keys)], "."))
	return Float64Conf
}

func GetIntConf(key string) int {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	IntConf := v.GetInt(strings.Join(keys[1:len(keys)], "."))
	return IntConf
}

func GetStringMapStringConf(key string) map[string]string {
	keys := strings.Split(key, ".")
	if len(key) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	StringMapStringConf := v.GetStringMapString(strings.Join(keys[1:len(keys)], "."))
	return StringMapStringConf
}

func GetStringSliceConf(key string) []string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	StringCliceConf := v.GetStringSlice(strings.Join(keys[1:len(keys)], "."))
	return StringCliceConf
}

func GetTimeConf(key string) time.Time {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return time.Now()
	}
	v := ViperConfMap[keys[0]]
	TimeConf := v.GetTime(strings.Join(keys[1:len(keys)], "."))
	return TimeConf
}

func GetDurationConf(key string) time.Duration {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	DurationConf := v.GetDuration(strings.Join(keys[1:len(keys)], "."))
	return DurationConf
}

//是否设置了key
func IsSetConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	IsSetConf := v.IsSet(strings.Join(keys[1:len(keys)], "."))
	return IsSetConf
}
