golang log library
[https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus "https://github.com/sirupsen/logrus")
1. 支持文件按天存储，基于logrus二次开发，
2. 增加日志自动按天分割
3. 支持指定路径存储
4. 支持到期自动删除

使用方法
安装go get github.com/ttllpp/golog
如果设置环境变量Environment=Development 默认输出文本拼接而不是json

```go
package main

import (
	log "github.com/ttllpp/golog"
)

func main() {
	log.SetLevel(log.DebugLevel)
	//如果不设置存储路径，默认直接输出
	log.SetPath("./", "test", 5)
	//使用方法1
	log.WithFields(log.Fields{
		"test": 111,
	}).Info("test")
	//使用方法2
	log.Info("test")
}

```


输出格式如下
[![](https://i.imgur.com/fV4dPcn.png)](https://i.imgur.com/fV4dPcn.png)

建议这样使用，可以详细输出具体字段
```go
log.WithFields(log.Fields{
		"test": 111,
		"name": "myname",
	}).Info("test")

```