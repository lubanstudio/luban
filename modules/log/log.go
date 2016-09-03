package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	loggers []*logger
)

// NewLogger initializes a new Logger with given buffer length, mode and configuration
// and add it to the loggers' queue.
func NewLogger(bufLen int64, mode, config string) {
	logger := newLogger(bufLen, mode, config)

	// This loop is to find the existing logger which has the same mode
	// and replace it with newly created one.
	isExist := false
	for i := range loggers {
		if loggers[i].mode == mode {
			isExist = true
			loggers[i] = logger
			break
		}
	}
	if !isExist {
		loggers = append(loggers, logger)
	}
}

// CloseLogger destroys logger with given mode and removes it from loggers' queue.
func CloseLogger(mode string) {
	for _, l := range loggers {
		if l.mode == mode {
			l.Close()
			break
		}
	}
}

func Trace(format string, v ...interface{}) {
	for _, l := range loggers {
		l.Trace(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	for _, l := range loggers {
		l.Info(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	for _, l := range loggers {
		l.Warn(format, v...)
	}
}

func Error(skip int, format string, v ...interface{}) {
	for _, l := range loggers {
		l.Error(skip, format, v...)
	}
}

func Fatal(skip int, format string, v ...interface{}) {
	for _, l := range loggers {
		l.Fatal(skip, format, v...)
	}
	Close()
	os.Exit(1)
}

// Close destroys all loggers in the queue.
func Close() {
	for _, l := range loggers {
		l.Close()
	}
}

// .___        __                 _____
// |   | _____/  |_  ____________/ ____\____    ____  ____
// |   |/    \   __\/ __ \_  __ \   __\\__  \ _/ ___\/ __ \
// |   |   |  \  | \  ___/|  | \/|  |   / __ \\  \__\  ___/
// |___|___|  /__|  \___  >__|   |__|  (____  /\___  >___  >
//          \/          \/                  \/     \/    \/

type Level int

const (
	TRACE Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger represents behaviors of a logger provider.
type Logger interface {
	Init(string) error
	Write(string, int, Level) error
	Destroy() error
	Flush()
}

type Factory func() Logger

var factories = make(map[string]Factory)

// Register registers given logger provider to adapters.
func Register(mode string, factory Factory) {
	if factory == nil {
		panic("log: register factory is nil for adapter \"" + mode + "\"")
	}
	if _, dup := factories[mode]; dup {
		panic("log: register called twice for adapter \"" + mode + "\"")
	}
	factories[mode] = factory
}

type logMsg struct {
	skip  int
	level Level
	msg   string
}

// Logger is default logger in beego application.
// it can contain several providers and log message into all providers.
type logger struct {
	mode    string
	adapter Logger
	level   Level
	msg     chan *logMsg
	quit    chan bool
}

// newLogger initializes and returns a new logger.
func newLogger(buffer int64, mode, config string) *logger {
	factory, ok := factories[mode]
	if !ok {
		panic("log: unknown adapter \"" + mode + "\" (forgotten register?)")
	}

	adapter := factory()
	if err := adapter.Init(config); err != nil {
		panic("log: fail to init adapter \"" + mode + "\": " + err.Error())
	}

	l := &logger{
		mode:    mode,
		adapter: adapter,
		msg:     make(chan *logMsg, buffer),
		quit:    make(chan bool),
	}

	go l.Start()
	return l
}

func (l *logger) write(skip int, level Level, msg string) error {
	if l.level > level {
		return nil
	}
	lm := &logMsg{
		skip:  skip,
		level: level,
	}

	// Only error information needs locate position for debugging.
	if lm.level >= ERROR {
		pc, file, line, ok := runtime.Caller(skip)
		if ok {
			// Get caller function name.
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn == nil {
				fnName = "?()"
			} else {
				fnName = strings.TrimLeft(filepath.Ext(fn.Name()), ".") + "()"
			}

			fileName := file
			if len(fileName) > 20 {
				fileName = "..." + fileName[len(fileName)-20:]
			}
			lm.msg = fmt.Sprintf("[%s:%d %s] %s", fileName, line, fnName, msg)
		} else {
			lm.msg = msg
		}
	} else {
		lm.msg = msg
	}
	l.msg <- lm
	return nil
}

// Start starts listening on read and quit chan.
func (l *logger) Start() {
	for {
		select {
		case bm := <-l.msg:
			if err := l.adapter.Write(bm.msg, bm.skip, bm.level); err != nil {
				fmt.Println("ERROR - unable to write:", err)
			}
		case <-l.quit:
			return
		}
	}
}

func (l *logger) Close() {
	l.quit <- true
	for {
		if len(l.msg) > 0 {
			bm := <-l.msg
			if err := l.adapter.Write(bm.msg, bm.skip, bm.level); err != nil {
				fmt.Println("ERROR, unable to WriteMsg:", err)
			}
		} else {
			break
		}
	}
	l.adapter.Flush()
	l.adapter.Destroy()
}

func (l *logger) Trace(format string, v ...interface{}) {
	msg := fmt.Sprintf("[TRACE] "+format, v...)
	l.write(0, TRACE, msg)
}

func (l *logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ INFO] "+format, v...)
	l.write(0, INFO, msg)
}

func (l *logger) Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ WARN] "+format, v...)
	l.write(0, WARN, msg)
}

func (l *logger) Error(skip int, format string, v ...interface{}) {
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	l.write(skip, ERROR, msg)
}

func (l *logger) Fatal(skip int, format string, v ...interface{}) {
	msg := fmt.Sprintf("[FATAL] "+format, v...)
	l.write(skip, FATAL, msg)
}
