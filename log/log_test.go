package log

import (
	"testing"
	"time"
)

//测试日志实例
func TestLogInstance(t *testing.T) {
	log := NewLoger()
	logConf := LogConfig{
		Level: "trace",
		FW: ConfFileWriter{
			On:              true,
			LogPath:         "./log_test.log",
			RotateLogPath:   "./log_test.log",
			WfLogPath:       "./log_test.wf.log",
			RotateWfLogPath: "./log_test.wf.log",
		},
		CW: ConfConsoleWriter{
			On:    true,
			Color: true,
		},
	}
	SetupLogInstanceWithConf(logConf, log)
	log.Info("testing")
	log.Close()
	time.Sleep(time.Second * 3)
}
