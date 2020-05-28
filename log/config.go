package log

import "errors"

//写文件
type ConfFileWriter struct {
	On              bool   `toml:"On"`
	LogPath         string `toml:"LogPath"` //
	RotateLogPath   string `toml:"RotateLogPath"`
	WfLogPath       string `toml:"WfLogPath"` //
	RotateWfLogPath string `toml:"RotateWfLogPath"`
}

//控制台配置
type ConfConsoleWriter struct {
	On    bool `toml:"On"`
	Color bool `toml:"Color"`
}

//日志配置
type LogConfig struct {
	Level string            `toml:"LogLevel"`
	FW    ConfFileWriter    `toml:"FileWriter"`    //文件
	CW    ConfConsoleWriter `toml:"ConsoleWriter"` //控制台
}

//使用file 和console的配置分别设置writer
func SetupLogInstanceWithConf(lc LogConfig, logger *Logger) (err error) {
	//FileWriter开启
	if lc.FW.On {
		if len(lc.FW.LogPath) > 0 {
			//生成FileWriter实例
			w := NewFileWriter()
			//设置文件名
			w.SetFileName(lc.FW.LogPath)
			//设置路径模式
			w.SetPathPattern(lc.FW.RotateLogPath)
			//设置级别
			w.SetLogLevelFloor(TRACE)
			if len(lc.FW.WfLogPath) > 0 {
				w.SetLogLevelCeil(INFO)
			} else {
				w.SetLogLevelCeil(ERROR)
			}
			//注册到[]Writer
			logger.Register(w)
		}

		if len(lc.FW.WfLogPath) > 0 {
			ww := NewFileWriter()
			ww.SetFileName(lc.FW.WfLogPath)
			ww.SetPathPattern(lc.FW.RotateWfLogPath)
			ww.SetLogLevelFloor(WARN)
			ww.SetLogLevelCeil(ERROR)
			//注册到[]Writer
			logger.Register(ww)
		}
	}

	//conslWriter开启
	if lc.CW.On {
		cw := NewConsoleWriter()
		cw.SetColor(lc.CW.Color)
		logger.Register(cw)
	}
	//日志的级别
	switch lc.Level {
	case "trace":
		logger.SetLevel(TRACE)
	case "debug":
		logger.SetLevel(DEBUG)
	case "info":
		logger.SetLevel(INFO)
	case "warning":
		logger.SetLevel(WARN)
	case "err":
		logger.SetLevel(ERROR)
	case "fatal":
		logger.SetLevel(FATAL)
	default:
		err = errors.New("Invalid log level")
	}
	return

}

//使用配置设置默认的log
func SetupDefaultWithConf(lc LogConfig) (err error) {
	defaultLoggerInit()
	return SetupLogInstanceWithConf(lc, logger_default)
}
