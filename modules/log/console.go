package log

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;34"), // Trace      blue
	newBrush("1;32"), // Info       green
	newBrush("1;33"), // Warn       yellow
	newBrush("1;31"), // Error      red
	newBrush("1;31"), // Fatal      red
}

// console implements Logger interface and writes messages to terminal.
type console struct {
	*log.Logger `json:"-"`
	Level       `json:"level"`
}

// newConsole creates a new console logger returning as Logger interface.
func newConsole() Logger {
	return &console{
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		Level:  TRACE,
	}
}

func (cw *console) Init(config string) error {
	return json.Unmarshal([]byte(config), cw)
}

func (cw *console) Write(msg string, skip int, level Level) error {
	if cw.Level > level {
		return nil
	}
	if runtime.GOOS == "windows" {
		cw.Logger.Println(msg)
	} else {
		cw.Logger.Println(colors[level](msg))
	}
	return nil
}

func (_ *console) Flush() {}

func (_ *console) Destroy() error {
	return nil
}

const CONSOLE = "console"

func init() {
	Register(CONSOLE, newConsole)
}
