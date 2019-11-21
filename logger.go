package cron


import (
	"io"
	"io/ioutil"
	"os"
	"log"
	"time"
	"strings"
)

var DefaultConf = map[string]string {
					"logfile" : "cron.log",
				}
var DefaultLogger = NewLogger(DefaultConf)
var DiscardConf = map[string]string {
					"discard" : "discard",
				}
var DiscardLogger = NewLogger(DiscardConf)

// Logger used in cron
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warning(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
}

type PrintLogger struct {
	logger interface { Printf(string,  ...interface{})}
}

func NewLogger(conf map[string]string) PrintLogger{
	var dst io.Writer
	var _, ok1 = conf["discard"]
	var logfile, ok2 = conf["logfile"]
	if ok1 {
		dst = ioutil.Discard
	} else if ok2{
		dst, _ = os.OpenFile(logfile, os.O_CREATE | os.O_RDWR | os.O_APPEND, 0666)
	} else {
		dst = os.Stdout
	}
	logger := log.New(dst, "", log.Ldate | log.Ltime)

	return PrintLogger{
		logger : logger,
	}
}

func (pl PrintLogger) Info (msg string, keysAndValues ...interface{}){
	keysAndValues = FormatTime(keysAndValues)
	spec := FormatString(len(keysAndValues))
	msg = "[INFO] " + msg
	pl.logger.Printf(spec, append([]interface{}{msg}, keysAndValues...)...)
}

func (pl PrintLogger) Warning (msg string, keysAndValues ...interface{}){
	keysAndValues = FormatTime(keysAndValues)
	spec := FormatString(len(keysAndValues))
	msg = "[WARNING] " + msg
	pl.logger.Printf(spec, append([]interface{}{msg}, keysAndValues...)...)
}

func (pl PrintLogger) Error (err error, msg string, keysAndValues ...interface{}){
	keysAndValues = FormatTime(keysAndValues)
	spec := FormatString(len(keysAndValues) + 2)
	msg = "[ERROR] " + msg
	pl.logger.Printf(spec, append([]interface{}{msg, "error", err}, keysAndValues...)...)
}

// FormatTime time.Time trans ti unix time
func FormatTime(arr []interface{}) []interface{} {
	var newarr []interface{}
	for _, v := range arr {
		if t, ok := v.(time.Time); ok {
			v = t.Unix()
		}
		newarr = append(newarr, v)
	}
	return newarr
}

// FormatString gen all args string format
func FormatString(argNum int) string {
	var buf strings.Builder
	buf.WriteString("%s")

	for i := 0; i < argNum / 2; i ++ {
		buf.WriteString(", %v=%v")
	}
	return buf.String()
}
