package log

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"

	"os"
	"path"
	"time"
)

//日志：写入文件

//路径变量表
var pathVariableTable map[byte]func(time *time.Time) int

//文件对象
type FileWriter struct {
	logLevelFloor int      //底
	logLevelCeil  int      //顶
	filename      string   //文件路径+名称
	pathFormat    string   //路径格式
	file          *os.File //存储日志的文件
	fileBufWriter *bufio.Writer
	actions       []func(time2 *time.Time) int //动作
	variables     []interface{}                //参数
}

//生成空文件对象
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (f *FileWriter) Init() error {
	return f.CreateLogFile()
}

func (f *FileWriter) SetFileName(filename string) {
	f.filename = filename
}

func (f *FileWriter) SetLogLevelFloor(floor int) {
	f.logLevelFloor = floor
}

func (f *FileWriter) SetLogLevelCeil(ceil int) {
	f.logLevelCeil = ceil
}

//创建日志文件
func (f *FileWriter) CreateLogFile() error {
	//0755->即用户具有读/写/执行权限，组用户和其它用户具有读写权限
	if err := os.MkdirAll(path.Dir(f.filename), 0755); err != nil {
		//如果创建文件出错
		if !os.IsExist(err) {
			return err
		}
	}
	//可对文件进行读写或者追加
	//0644->即用户具有读写权限，组用户和其它用户具有只读权限
	if file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_APPEND, 0644); err != nil {
		return err
	} else {
		f.file = file
	}
	//为实例赋值 文件及大小
	if f.fileBufWriter = bufio.NewWriterSize(f.file, 8192); f.fileBufWriter == nil {
		return errors.New("初始化filebufWriter失败")
	}
	return nil
}

//设置组路径模式
func (f *FileWriter) SetPathPattern(pattern string) error {
	//包含几个%
	n := 0
	for _, c := range pattern {
		if c == '%' {
			n++
		}
	}
	if n == 0 {
		f.pathFormat = pattern
		return nil
	}
	f.actions = make([]func(time2 *time.Time) int, 0, n)
	f.variables = make([]interface{}, n, n)
	tmp := []byte(pattern)

	variable := 0
	for _, c := range tmp {
		if variable == 1 {
			act, ok := pathVariableTable[c]
			if !ok {
				return errors.New("无效的模式(" + pattern + ")")
			}
			f.actions = append(f.actions, act)
			variable = 0
			continue
		}
		if c == '%' {
			variable = 1
		}
	}

	for i, act := range f.actions {
		now := time.Now()
		f.variables[i] = act(&now)
	}
	//设置时间格式
	f.pathFormat = convertPatternToFormat(tmp)
	return nil
}

func (f *FileWriter) Rotate() error {
	now := time.Now()
	v := 0
	rotate := false
	old_variables := make([]interface{}, len(f.variables))
	copy(old_variables, f.variables)

	for i, act := range f.actions {
		//执行处理时间的函数
		v = act(&now)
		if v != f.variables[i] {
			f.variables[i] = v
			rotate = true
		}

	}
	if rotate == false {
		return nil
	}
	if f.fileBufWriter != nil {
		//刷新会将所有缓冲的数据写入基础io.Writer
		if err := f.fileBufWriter.Flush(); err != nil {
			return nil
		}
	}
	//已经设置的存储日志的文件及大小
	if f.file != nil {
		//将文件以pattern形式改名
		filePath := fmt.Sprintf(f.pathFormat, old_variables...)

		if err := os.Rename(f.filename, filePath); err != nil {
			return err
		}
		//关闭文件
		if err := f.file.Close(); err != nil {
			return err
		}

	}
	//生成log文件
	return f.CreateLogFile()
}

//将记录写入文件
func (f *FileWriter) Write(r *Record) error {
	//如果recod的级别不在允许额范围内
	if r.level < f.logLevelFloor || r.level > f.logLevelCeil {
		return nil
	}
	if f.fileBufWriter == nil {
		return errors.New("not a opened file")
	}
	if _, err := f.fileBufWriter.WriteString(r.String()); err != nil {
		return err
	}
	return nil
}

func (f *FileWriter) Flush() error {
	if f.fileBufWriter != nil {
		return f.fileBufWriter.Flush()
	}
	return nil
}

//设置时间：
func convertPatternToFormat(pattern []byte) string {

	pattern = bytes.Replace(pattern, []byte("%Y"), []byte("%d"), -1)
	pattern = bytes.Replace(pattern, []byte("%M"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%D"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%H"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%m"), []byte("%02d"), -1)
	return string(pattern)
}

func init() {
	pathVariableTable = make(map[byte]func(*time.Time) int, 5)
	pathVariableTable['Y'] = getYear
	pathVariableTable['M'] = getMonth
	pathVariableTable['D'] = getDay
	pathVariableTable['H'] = getHour
	pathVariableTable['m'] = getMin
}

func getYear(now *time.Time) int {
	return now.Year()
}

func getMonth(now *time.Time) int {
	return int(now.Month())
}

func getDay(now *time.Time) int {
	return now.Day()
}

func getHour(now *time.Time) int {
	return now.Hour()
}

func getMin(now *time.Time) int {
	return now.Minute()
}
