golang log library

1. 支持文件按天存储
2. 增加日志自动按天分割
3. 支持指定路径存储
4. 支持到期自动删除
5. 支持队列(需自己扩展，在messageQueue.go里)
6. 按等级划分
7. 根据当前环境选择是否打印日志到终端

使用方法
安装go get github.com/ttllpp/golog
如果设置环境变量RUNMODE=dev 默认输出到终端

```go
package main

import (
	log "github.com/ttllpp/golog"
)

func main() {
	//初始化日志
	log.GeneralInit()
	//或者自定义初始化
	//
	//Start(FatalLevel, AlsoStdout, LogFilePath(saveLogFilePath), EveryDay,  Appid(appid), MessageQueueInstance, FatalMessageQueueLevel)
	log.Debugln("aaa")
	log.Debugf("bb")
}

```


输出格式如下
```
2019/11/22 14:57:06 DEBUG [golog.TestLog] (/xxx/log_test.go:10) - aaa
2019/11/22 14:57:06 DEBUG [golog.TestLog] (/xxx/log_test.go:10) - bb


```