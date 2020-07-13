package golog

//发布日志类消息

// "errors"

type messageQueue struct {
	// messagequeue.Base
}

type PushType struct {
	Appid    int
	LogId    string
	LogTime  string
	LogText  string
	LogLevel LogLevel
}

func newMessageQueueInstance() *messageQueue {
	messageQueue := &messageQueue{}
	return messageQueue
}

func (this *messageQueue) publishLog(content PushType) error {
	return nil
}
