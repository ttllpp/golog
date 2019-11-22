package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Start(DebugLevel, LogFilePath("./temp/"), DebugLevel, EveryMinute, AlsoStdout)

	Debugln("debug")

}