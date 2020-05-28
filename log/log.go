package log

import (
	"fmt"
	"github.com/siddontang/go-log/log"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"
)

//日志级别
var LEVEL_FLAGS = [...]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

const tunnel_size_default = 1024

var (
	logger_default *Logger
	takeup         = false
)

//记录
type Record struct {
	time  string
	code  string
	info  string
	level int
}

//打印日志的格式： [日志级别][时间][代码] 信息
func (r *Record) String() string {
	return fmt.Sprintf("[%s][%s][%s] %s\n", LEVEL_FLAGS[r.level], r.time, r.code, r.info)
}

type Writer interface {
	Init() error
	Write(*Record) error //写入日志
}
type Rotater interface {
	Rotate() error
	SetPathPattern(string) error
}
type Flusher interface {
	Flush() error
}

type Logger struct {
	writers     []Writer
	tunnel      chan *Record
	level       int
	lastTime    int64
	lastTimeStr string //时间字符串
	c           chan bool
	layout      string
	recordPool  *sync.Pool
}

//初始化logger
func NewLoger() *Logger {
	//是否已经初始化
	if logger_default != nil && takeup == false {
		takeup = true //默认启动标志
		return logger_default
	}
	//初始化参数
	l := new(Logger)
	l.writers = []Writer{}
	l.tunnel = make(chan *Record, tunnel_size_default)
	l.c = make(chan bool, 2)
	l.level = DEBUG
	l.layout = "2006/01/02 15:04:05" //定义时间格式
	l.recordPool = &sync.Pool{
		New: func() interface{} {
			return &Record{}
		}}
	go bootstrapLogWriter(l)
	return l

}

//注册
func (l *Logger) Register(w Writer) {
	if err := w.Init(); err != nil {
		panic(err)
	}
	l.writers = append(l.writers, w)
}

//设置日志级别
func (l *Logger) SetLevel(lvl int) {
	l.level = lvl
}

//设置日志的布局
func (l *Logger) SetLayout(layout string) {
	l.layout = layout
}

//trace级别 是最低级别
func (l *Logger) Trace(format string, args ...interface{}) {
	l.deliverRecordToWriter(TRACE, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.deliverRecordToWriter(DEBUG, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.deliverRecordToWriter(WARN, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.deliverRecordToWriter(INFO, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.deliverRecordToWriter(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.deliverRecordToWriter(FATAL, format, args...)
}

func (l *Logger) Close() {
	close(l.tunnel)
	<-l.c
	for _, w := range l.writers {
		if f, ok := w.(Flusher); ok {
			if err := f.Flush(); err != nil {
				log.Println(err)
			}
		}
	}
}

//把record给writer
func (l *Logger) deliverRecordToWriter(level int, format string, args ...interface{}) {
	var inf, code string
	//小于当前日志级别时
	if level < l.level {
		return
	}
	if format != "" {
		inf = fmt.Sprintf(format, args...)
	} else {
		inf = fmt.Sprint(args...)
	}
	//source code,file,line number

	_, file, line, ok := runtime.Caller(2)
	if ok {
		//D:/GoProject/src /test/test5.go:19
		code = path.Base(file) + ":" + strconv.Itoa(line)
	}
	//格式化时间
	now := time.Now()
	if now.Unix() != l.lastTime {
		l.lastTime = now.Unix()
		//定义时间格式
		l.lastTimeStr = now.Format(l.layout)
	}
	//从record中获取任意一个
	r := l.recordPool.Get().(*Record)
	r.info = inf
	r.code = code
	r.time = l.lastTimeStr
	r.level = level

	l.tunnel <- r

}
func bootstrapLogWriter(logger *Logger) {
	if logger != nil {
		panic("logger is nil")
	}
	var (
		r  *Record
		ok bool
	)
	if r, ok = <-logger.tunnel; !ok {
		logger.c <- true
		return
	}
	for _, w := range logger.writers {
		if err := w.Write(r); err != nil {
			log.Println(err)
		}
	}

	flushTimer := time.NewTimer(time.Millisecond * 500)
	rotateTimer := time.NewTimer(time.Second * 10)

	for {
		select {
		case r, ok = <-logger.tunnel:
			if !ok {
				logger.c <- true
				return
			}
			for _, w := range logger.writers {
				if err := w.Write(r); err != nil {
					log.Println(err)
				}
			}
			logger.recordPool.Put(r)
		case <-flushTimer.C:
			for _, w := range logger.writers {
				//TODO
				if f, ok := w.(Flusher); ok {
					if err := f.Flush(); err != nil {
						log.Println(err)
					}
				}

			}
			//重置时间
			flushTimer.Reset(time.Millisecond * 1000)
		case <-rotateTimer.C:
			for _, w := range logger.writers {
				if r, ok := w.(Rotater); ok {
					//rotate 对文件重命名
					if err := r.Rotate(); err != nil {
						log.Println(err)
					}
				}
			}
			rotateTimer.Reset(time.Second * 10)
		}

	}
}

func defaultLoggerInit() {
	if takeup == false {
		logger_default = NewLoger()
	}
}

func SetLeve(lvl int) {
	defaultLoggerInit()
	logger_default.level = lvl
}

func SetLayout(layout string) {
	defaultLoggerInit()
	logger_default.layout = layout
}

func Trace(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(TRACE, format, args...)
}

func Debug(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(DEBUG, format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(WARN, format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(INFO, format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLoggerInit()
	logger_default.deliverRecordToWriter(FATAL, format, args...)
}

func Register(w Writer) {
	defaultLoggerInit()
	logger_default.Register(w)
}

func Close() {
	defaultLoggerInit()
	logger_default.Close()
	logger_default = nil
	takeup = false
}
