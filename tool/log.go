package tool

import (
	"fmt"
	dlog "lib/log"
	"strings"
)

//常用请求常量定义
const (
	DLTagUndefind      = "_undef"              //未定义
	DLTagMysqlFailed   = "_com_mysql_failure"  //MySQL连接失败
	DLTagRedisFailed   = "_com_redis_failure"  //redis连接失败
	DLTagMysqlSuccess  = "_com_mysql_success"  //MySQL连接成功
	DLTagredisSuccess  = "_com_redis_success"  //redis连接成功
	DLTagThriftFailed  = "_com_thrift_failure" //thrift连接失败
	DLTagThriftSuccess = "_com_thrift_success" //thrift连接成功
	DLTagHTTPSuccess   = "_com_http_success"   //http连接成功
	DLTagHTTPFailed    = "_com_http_failure"   //http连接失败
	DLTagRequestIn     = "_com_request_in"     //请求
	DLTagRequstOut     = "_com_request_out"    //响应

)

const (
	_dlTag          = "dltag"
	_traceId        = "traceid"
	_spanId         = "spanid"
	_childSpanId    = "cspanid"
	_dlTagBizPrefix = "_com_"
	_dlTagBizUndef  = "_com_undef"
)

var Log *Logger

type Logger struct {
}

type Trace struct {
	TraceId     string
	SpanId      string
	Caller      string
	SrcMethod   string
	HintCode    int64
	HintContent string
}

type TraceContext struct {
	Trace
	CSpanId string
}

func (l *Logger) TagInfo(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Info(parseParams(m))
}

func (l *Logger) TagWarn(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Warn(parseParams(m))
}

func (l *Logger) TagError(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Error(parseParams(m))
}

func (l *Logger) TagTrace(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Trace(parseParams(m))
}

func (l *Logger) TagDebug(trace *TraceContext, dltag string, m map[string]interface{}) {
	m[_dlTag] = checkDLTag(dltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	dlog.Debug(parseParams(m))
}

func (l *Logger) Close() {
	dlog.Close()
}

//生成业务dltag
func CreateBizDLTag(tagName string) string {
	if tagName == "" {
		return _dlTagBizUndef
	}
	return _dlTagBizPrefix + tagName
}

//校验dltag的合法性
func checkDLTag(dltag string) string {
	if strings.HasPrefix(dltag, _dlTagBizPrefix) {
		return dltag
	}

	if strings.HasPrefix(dltag, "_com_") {
		return dltag
	}
	//未定时
	if dltag == DLTagUndefind {
		return dltag
	}
	return dltag
}

//map格式化为string
func parseParams(m map[string]interface{}) string {
	//格式化后的字符串
	var dltag string = "_undef"
	if _dltag, _have := m["dltag"]; _have {
		if _val, _ok := _dltag.(string); _ok {
			dltag = _val
		}
	}

	for _key, _val := range m {
		if _key == "dltag" {
			continue
		}
		dltag = dltag + "||" + fmt.Sprintf("%v=%+v", _key, _val)
	}
	dltag = strings.Trim(fmt.Sprintf("%q", dltag), "\"")
	return dltag
}
