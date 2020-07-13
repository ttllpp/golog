package golog

import (
	"testing"
)

func TestLog(t *testing.T) {
	Start(DebugLevel, AlsoStdout)
	Start(InfoLevel, AlsoStdout, LogFilePath("./temp"), EveryDay)

	// GeneralInit()
	Debugf("debug")
	Infof("info")
}
