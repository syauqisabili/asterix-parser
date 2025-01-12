package pkg

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/charmbracelet/log"
)

// Logger is a wrapper around the charmbracelet/log.Logger
var (
	logger *log.Logger
	once   sync.Once
)

func LogInfo(msg interface{}, keyvals ...interface{}) {
	logger.Info(msg, keyvals...)
}

func LogDebug(msg interface{}, keyvals ...interface{}) {
	// Get the current caller's information
	pc, file, line, ok := runtime.Caller(1) // 1 means the caller of this function
	if !ok {
		file = "unknown"
		line = 0
	}

	// Get only the base name of the file
	file = filepath.Base(file) // Extract the base file name

	// Get the function name
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName) // Only keep the function name

	// Only debug show file:func:line
	logger.Debug(fmt.Sprintf("<%s:%s:%d>: %s", file, funcName, line, msg), keyvals...)
}

func LogWarn(msg interface{}, keyvals ...interface{}) {
	logger.Warn(msg, keyvals...)
}

func LogError(msg interface{}, keyvals ...interface{}) {
	logger.Error(msg, keyvals...)
}

func LogFatal(msg interface{}, keyvals ...interface{}) {
	logger.Fatal(msg, keyvals...)
}
