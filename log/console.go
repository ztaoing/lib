package log

import (
	"fmt"
	"os"
)

type colorRecord Record

//根据不同的等级设置输出样式
func (r *colorRecord) String() string {
	switch r.level {
	case TRACE:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[34m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	case DEBUG:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[34m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case INFO:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[32m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)

	case WARN:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[33m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	case ERROR:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[33m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	case FATAL:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[33m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LEVEL_FLAGS[r.level], r.code, r.info)
	}
	return ""
}

type ConsoleWriter struct {
	color bool
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

//对写入到控制台的数据采用样式
func (w *ConsoleWriter) Write(r *Record) error {
	if w.color {
		//将record转换为colorRecord
		fmt.Fprint(os.Stdout, ((*colorRecord)(r)).String())
	} else {
		//无样式
		fmt.Fprint(os.Stdout, r.String())
	}
	return nil
}

func (w *ConsoleWriter) Init() error {
	return nil
}

//是否启动输出样式
func (w *ConsoleWriter) SetColor(c bool) {
	w.color = c
}
