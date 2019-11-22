package log

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

//初始化方法需要调用装饰器去构造，用哪种级别的就构造哪种的，没构造的就不会起作用
// 构造debug
// loger.Start(loger.DebugLevel, loger.AlsoStdout)
// 构造info
// loger.Start(loger.InfoLevel, loger.AlsoStdout, loger.LogFilePath(saveLogFilePath), loger.EveryDay)
// 构造warn
// loger.Start(loger.WarnLevel, loger.AlsoStdout)
// 构造error
// loger.Start(loger.ErrorLevel, loger.AlsoStdout, loger.LogFilePath(saveLogFilePath), loger.EveryDay, loger.Appid(int(id.(int64))), loger.MessageQueueInstance, loger.ErrorMessageQueueLevel)
// 构造fatal
// loger.Start(loger.FatalLevel, loger.AlsoStdout, loger.LogFilePath(saveLogFilePath), loger.EveryDay, loger.Appid(int(id.(int64))), loger.MessageQueueInstance, loger.FatalMessageQueueLevel)
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// ++++我是分割线++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// 常用方法
// Debugf(format string, v ...interface{})
// Infof(format string, v ...interface{})
// Warnf(format string, v ...interface{})
// Errorf(format string, v ...interface{})
// Fatalf(format string, v ...interface{})
// Debugln(v ...interface{})
// Infoln(v ...interface{})
// Warnln(v ...interface{})
// Errorln(v ...interface{})
// Fatalln(v ...interface{})

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	started             int32
	debugLoggerInstance Logger
	infoLoggerInstance  Logger
	warnLoggerInstance  Logger
	errorLoggerInstance Logger
	fatalLoggerInstance Logger
	Expired             int = 7 //日志有效期
	tagName                 = map[LogLevel]string{
		DEBUG: "DEBUG", //指出细粒度信息事件对调试应用程序是非常有帮助的，主要用于开发过程中打印一些运行信息
		INFO:  "INFO",  //消息在粗粒度级别上突出强调应用程序的运行过程。打印一些你感兴趣的或者重要的信息，这个可以用于生产环境中输出程序运行的一些重要信息，但是不能滥用，避免打印过多的日志。
		WARN:  "WARN",  //应该是这个时候进行一些修复性的工作，应该还可以把系统恢复到正常状态中来，系统应该可以继续运行下去
		ERROR: "ERROR", //指出虽然发生错误事件，但仍然不影响系统的继续运行，如果不想输出太多的日志，可以使用这个级别。
		FATAL: "FATAL", //可以肯定这种错误已经无法修复，可以肯定必然会越来越乱
	}
)

//启动日志
func Start(decorators ...func(Logger) Logger) Logger {
	loggerInstance := Logger{}
	for _, decorator := range decorators {
		loggerInstance = decorator(loggerInstance)
	}
	var logger *log.Logger
	var segment *logSegment
	if loggerInstance.logPath != "" {
		segment = newLogSegment(loggerInstance.unit, loggerInstance.logPath, loggerInstance.level)
	}
	//未设置队列接收的时候，设置messageQueueLevel为最大值
	if loggerInstance.messageQueue == nil {
		loggerInstance.messageQueueLevel = LogLevel(^uint(0) >> 1)
	}
	if segment != nil {
		logger = log.New(segment, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	loggerInstance.logger = logger
	switch loggerInstance.level {
	case DEBUG:
		debugLoggerInstance = loggerInstance
	case INFO:
		infoLoggerInstance = loggerInstance
	case WARN:
		warnLoggerInstance = loggerInstance
	case ERROR:
		errorLoggerInstance = loggerInstance
	case FATAL:
		fatalLoggerInstance = loggerInstance
	}
	return loggerInstance
}

func (l Logger) Stop() {
	if l.printStack {
		traceInfo := make([]byte, 1<<16)
		n := runtime.Stack(traceInfo, true)
		l.logger.Printf("%s", traceInfo[:n])
		if l.isStdout {
			log.Printf("%s", traceInfo[:n])
		}
	}
	if l.segment != nil {
		l.segment.Close()
	}
	l.segment = nil
	l.logger = nil
}

// logSegment 实现 io.Writer 的Write方法
type logSegment struct {
	unit         time.Duration
	logPath      string
	logFile      *os.File
	timeToCreate <-chan time.Time
	loggerLevel  LogLevel
}

func getNowTime() time.Time {
	cstSh := getTimeLocation()
	return time.Now().In(cstSh)
}

func getTimeLocation() *time.Location {
	cstSh, _ := time.LoadLocation("Asia/Shanghai")
	return cstSh
}

func newLogSegment(unit time.Duration, logPath string, level LogLevel) *logSegment {
	now := getNowTime()
	if logPath != "" {
		err := os.MkdirAll(logPath, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil
		}
		nowTime := now
		cstSh := getTimeLocation()
		name := getLogFileName(nowTime, level)
		tempLogName := path.Join(logPath, name)
		if fileInfo, err := os.Stat(tempLogName); os.IsExist(err) {
			if fileInfo.ModTime().In(cstSh).Day() != nowTime.Day() {
				os.Remove(tempLogName)
			}
		}
		logFile, err := os.OpenFile(tempLogName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			if os.IsNotExist(err) {
				logFile, err = os.Create(path.Join(logPath, name))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return nil
				}
			} else {
				fmt.Fprintln(os.Stderr, err)
				return nil
			}
		}
		next := now.Truncate(unit).Add(unit)
		var timeToCreate <-chan time.Time
		if unit == time.Hour*24 || unit == time.Hour || unit == time.Minute {
			timeToCreate = time.After(next.Sub(time.Now()))
		}
		logSegment := &logSegment{
			unit:         unit,
			logPath:      logPath,
			logFile:      logFile,
			timeToCreate: timeToCreate,
			loggerLevel:  level,
		}
		return logSegment
	}
	return nil
}

func (ls *logSegment) Write(p []byte) (int, error) {
	if ls.timeToCreate != nil && ls.logFile != os.Stdout && ls.logFile != os.Stderr {
		select {
		case current := <-ls.timeToCreate:
			ls.Close()
			name := getLogFileName(current, ls.loggerLevel)
			var err error
			ls.logFile, err = os.Create(path.Join(ls.logPath, name))
			if err != nil {
				// 如果不能写文件的话，直接输出到系统错误
				fmt.Fprintln(os.Stderr, err)
				ls.logFile = os.Stderr
			} else {
				next := current.Truncate(ls.unit).Add(ls.unit)
				ls.timeToCreate = time.After(next.Sub(time.Now()))
			}
		default:
			// do nothing
		}
	}
	return ls.logFile.Write(p)
}

func (ls *logSegment) Close() {
	ls.logFile.Close()
}

func getLogFileName(now time.Time, level LogLevel) string {
	day := now.Day()
	return fmt.Sprintf("%d.%s.log", day%Expired, tagName[level])
}

type Logger struct {
	logger            *log.Logger
	level             LogLevel
	messageQueueLevel LogLevel
	messageQueue      *messageQueue
	segment           *logSegment
	stopped           int32
	logPath           string
	unit              time.Duration
	isStdout          bool
	printStack        bool
	appid             int
}

func (l Logger) doPrintf(level LogLevel, format string, v ...interface{}) {
	if l.logger == nil {
		return
	}
	if level >= l.messageQueueLevel || level >= l.level {
		funcName, fileName, lineNum := getRuntimeInfo()
		format = fmt.Sprintf("%5s [%s] (%s:%d) - %s", tagName[level], funcName, fileName, lineNum, format)
		logText := fmt.Sprintf(format, v...)
		if level >= l.messageQueueLevel {
			pushType := PushType{
				Appid:    l.appid,
				LogTime:  getNowTime().Format("2006-01-02 15:04:05"),
				LogText:  logText,
				LogLevel: level,
			}
			l.messageQueue.publishLog(pushType)
		}
		if level >= l.level {
			l.logger.Print(logText)
			if l.isStdout {
				log.Print(logText)
			}
			if level == FATAL {
				// os.Exit(1)
			}
		}
	}
}

func (l Logger) doPrintln(level LogLevel, v ...interface{}) {
	if l.logger == nil {
		return
	}
	if level >= l.messageQueueLevel || level >= l.level {
		funcName, fileName, lineNum := getRuntimeInfo()
		prefix := fmt.Sprintf("%5s [%s] (%s:%d) - ", tagName[level], path.Base(funcName), fileName, lineNum)
		value := fmt.Sprintf("%s%s", prefix, fmt.Sprintln(v...))
		if level >= l.messageQueueLevel {
			pushType := PushType{
				Appid:    l.appid,
				LogTime:  getNowTime().Format("2006-01-02 15:04:05"),
				LogText:  value,
				LogLevel: level,
			}
			l.messageQueue.publishLog(pushType)
		}
		if level >= l.level {
			l.logger.Print(value)
			if l.isStdout {
				log.Print(value)
			}
			if level == FATAL {
				// os.Exit(1)
			}
		}
	}
}

func stack() string {
	tracedata := ""
	//trace日志最大深度为15
	pc := make([]uintptr, 15)
	n := runtime.Callers(5, pc)
	for i := 0; i < n; i++ {
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		tracedata += fmt.Sprintf("%s:%d %s\n", file, line, f.Name())
	}
	return tracedata
}

func getRuntimeInfo() (string, string, int) {
	pc, fn, ln, ok := runtime.Caller(3)
	if !ok {
		fn = "???"
		ln = 0
	}
	function := "???"
	caller := runtime.FuncForPC(pc)
	if caller != nil {
		function = caller.Name()
	}
	return function, fn, ln
}

//check meeesageQueue
func checkMessageQueue(l Logger) error {
	if l.messageQueue == nil {
		panic("定义MessageQueueLevel前先初始化messageQueue")
	}
	return nil
}

//这个应该由各个框架在自己需要recover地方自己实现，并记录所需要的信息
//log这里只是简单实现的PanicRecover 具体业务场景可以自己根据需求实现
func PanicRecover() {
	if err := recover(); err != nil {
		var stack string
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			stack = stack + fmt.Sprintln(fmt.Sprintf("%s:%d", file, line))
		}
		log.Fatalf("爆表堆栈信息:%v", stack)
	}
}

// 常用初始化方式example
// debug日志只是输出界面
// info,warn,error,fatal日志输出文件
// error,fatal日志上报kafka
func GeneralInit() {
	var appid int = 1
	saveLogFilePath := "./log/"
	
	if os.Getenv("RUNMODE") == "dev" {
		//debug日志
		Start(DebugLevel, AlsoStdout)
	}
	//info日志
	Start(InfoLevel, AlsoStdout, LogFilePath(saveLogFilePath), EveryDay)
	//warn日志
	Start(WarnLevel, AlsoStdout, LogFilePath(saveLogFilePath), EveryDay)
	//error日志
	Start(ErrorLevel, AlsoStdout, LogFilePath(saveLogFilePath), EveryDay, Appid(appid), MessageQueueInstance, ErrorMessageQueueLevel)
	//fatal日志
	Start(FatalLevel, AlsoStdout, LogFilePath(saveLogFilePath), EveryDay,  Appid(appid), MessageQueueInstance, FatalMessageQueueLevel)
}

//
//
// 下面是装饰器
//
//
//日志打印和输出到文件的等级
func DebugLevel(l Logger) Logger {
	l.level = DEBUG
	return l
}

func InfoLevel(l Logger) Logger {
	l.level = INFO
	return l
}

func WarnLevel(l Logger) Logger {
	l.level = WARN
	return l
}

func ErrorLevel(l Logger) Logger {
	l.level = ERROR
	return l
}

func FatalLevel(l Logger) Logger {
	l.level = FATAL
	return l
}

//日志输出到消息队列的日志等级装饰器
func DebugMessageQueueLevel(l Logger) Logger {
	checkMessageQueue(l)
	l.messageQueueLevel = DEBUG
	return l
}

func InfoMessageQueueLevel(l Logger) Logger {
	checkMessageQueue(l)
	l.messageQueueLevel = INFO
	return l
}

func WarnMessageQueueLevel(l Logger) Logger {
	checkMessageQueue(l)
	l.messageQueueLevel = WARN
	return l
}

func ErrorMessageQueueLevel(l Logger) Logger {
	checkMessageQueue(l)
	l.messageQueueLevel = ERROR
	return l
}

func FatalMessageQueueLevel(l Logger) Logger {
	checkMessageQueue(l)
	l.messageQueueLevel = FATAL
	return l
}

//消息队列初始化
func MessageQueueInstance(l Logger) Logger {
	l.messageQueue = newMessageQueueInstance()
	return l
}

//日志输出路径
func LogFilePath(p string) func(Logger) Logger {
	return func(l Logger) Logger {
		l.logPath = p
		return l
	}
}

//appid 默认0
func Appid(appid int) func(Logger) Logger {
	return func(l Logger) Logger {
		l.appid = appid
		return l
	}
}

//每天1个日志文件
func EveryDay(l Logger) Logger {
	l.unit = time.Hour * 24
	return l
}

//每小时1个日志文件
func EveryHour(l Logger) Logger {
	l.unit = time.Hour
	return l
}

//每分钟1个日志文件
func EveryMinute(l Logger) Logger {
	l.unit = time.Minute
	return l
}

//总是输出到视窗
func AlsoStdout(l Logger) Logger {
	l.isStdout = true
	return l
}

//Stop的时候打堆栈
func PrintStack(l Logger) Logger {
	l.printStack = true
	return l
}

//
//
// 下面是外部打日志的方法
//
//
func Debugf(format string, v ...interface{}) {
	debugLoggerInstance.doPrintf(DEBUG, format, v...)
}

func Infof(format string, v ...interface{}) {
	infoLoggerInstance.doPrintf(INFO, format, v...)
}

func Warnf(format string, v ...interface{}) {
	warnLoggerInstance.doPrintf(WARN, format, v...)
}

func Errorf(format string, v ...interface{}) {
	errorLoggerInstance.doPrintf(ERROR, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	fatalLoggerInstance.doPrintf(FATAL, format, v...)
}

func Debugln(v ...interface{}) {
	debugLoggerInstance.doPrintln(DEBUG, v...)
}

func Infoln(v ...interface{}) {
	infoLoggerInstance.doPrintln(INFO, v...)
}

func Warnln(v ...interface{}) {
	warnLoggerInstance.doPrintln(WARN, v...)
}

func Errorln(v ...interface{}) {
	errorLoggerInstance.doPrintln(ERROR, v...)
}

func Fatalln(v ...interface{}) {
	fatalLoggerInstance.doPrintln(FATAL, v...)
}

//纯打印，无任何作用，为了兼容以前的log.Println
func Println(v ...interface{}) {
	log.Println(v...)
}
