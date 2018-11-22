package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type logFile struct {
	logfile  *os.File
	fileName string
	path     string
	Log      *logrus.Logger
}
type Fields = logrus.Fields

const (
	DebugLevel = logrus.DebugLevel
	PanicLevel = logrus.PanicLevel
	FatalLevel = logrus.FatalLevel
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	TraceLevel = logrus.TraceLevel
)

var logOb *logFile

func init() {
	logOb = &logFile{
		Log: logrus.New(),
	}
	var environment string
	environment = os.Getenv("Environment")
	if environment == "Development" {
		logOb.Log.Formatter = &logrus.TextFormatter{}
	} else {
		logOb.Log.Formatter = &logrus.JSONFormatter{}
	}
	logOb.Log.SetLevel(logrus.DebugLevel)
}

func SetPath(path, filename string, expiredDay int) {
	logOb.path = path
	logOb.fileName = filename
	openFile()
	if expiredDay <= 0 {
		expiredDay = 7
	}
	go setLogExpiredDay(expiredDay)
}

func SetLevel(level logrus.Level) {
	logOb.Log.SetLevel(level)
}

func setLogExpiredDay(day int) {
	for {
		path := logOb.path + time.Now().AddDate(0, 0, -day).Format("20060102") + "/"
		filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
			if nil == fi {
				return err
			}
			if !fi.IsDir() {
				return nil
			}
			os.RemoveAll(path)
			return nil
		})
		<-time.After(24 * time.Hour)
	}
}

func openFile() {
	pid := strconv.Itoa(os.Getpid())
	dir := logOb.path + time.Now().Format("20060102") + "/"
	exist, _ := pathExists(dir)
	if exist == false {
		os.MkdirAll(dir, os.ModePerm)
	}
	logfile, _ := os.OpenFile(dir+logOb.fileName+"_"+time.Now().Format("150405")+"_"+pid, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	logOb.Log.Out = logfile
	logOb.logfile.Close()
	logOb.logfile = logfile
	go dayRotatelogs()
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func dayRotatelogs() {
	timeStr := time.Now().Format("2006-01-02")
	todayTime, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	interval := todayTime.AddDate(0, 0, 1).Unix() - time.Now().Unix()
	<-time.After(time.Duration(interval) * time.Second)
	openFile()
}

func Debug(v interface{}) {
	logOb.Log.Debug(v)
}
func Info(v interface{}) {
	logOb.Log.Info(v)
}
func Warn(v interface{}) {
	logOb.Log.Warn(v)
}
func Error(v interface{}) {
	logOb.Log.Error(v)
}
func Fatal(v interface{}) {
	logOb.Log.Fatal(v)
}
func Panic(v interface{}) {
	logOb.Log.Panic(v)
}
func WithFields(fields Fields) *logrus.Entry {
	return logOb.Log.WithFields(fields)
}

