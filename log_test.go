package log

import (
	"testing"
)

func Test_NewLog(t *testing.T) {
	SetLevel(DebugLevel)
	SetPath("./", "test", 5)
	WithFields(Fields{
		"test": 111,
	}).Info("test")
	Info("test")
}

