package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// file implements Logger interface.
// It writes messages by lines limit, file size limit, or time frequency.
type file struct {
	*log.Logger
	writer *muxWriter

	Filename string `json:"filename"`

	Rotate bool `json:"rotate"`

	// Rotate at lines
	Maxlines     int `json:"maxlines"`
	currentlines int

	// Rotate at size
	Maxsize     int `json:"maxsize"`
	currentSize int

	// Rotate daily
	Daily    bool  `json:"daily"`
	Maxdays  int64 `json:"maxdays"`
	opendate int

	startLock sync.Mutex // Only one log can write to the file

	Level `json:"level"`
}

// an *os.File writer with locker.
type muxWriter struct {
	sync.Mutex
	file *os.File
}

// write to os.File.
func (mw *muxWriter) Write(b []byte) (int, error) {
	mw.Lock()
	defer mw.Unlock()
	return mw.file.Write(b)
}

// set os.File in writer.
func (mw *muxWriter) SetFd(file *os.File) {
	if mw.file != nil {
		mw.file.Close()
	}
	mw.file = file
}

// newFile creates a file logger returning as Logger interface.
func newFile() Logger {
	w := &file{
		Filename: "",
		Maxlines: 1000000,
		Maxsize:  1 << 28, //256 MB
		Daily:    true,
		Maxdays:  7,
		Rotate:   true,
		Level:    TRACE,
	}
	// use MuxWriter instead direct use os.File for lock write when rotate
	w.writer = new(muxWriter)
	// set MuxWriter as Logger's io.Writer
	w.Logger = log.New(w.writer, "", log.Ldate|log.Ltime)
	return w
}

// start file logger. create log file and set to locker-inside file writer.
func (w *file) start() error {
	fd, err := w.createLogFile()
	if err != nil {
		return err
	}
	w.writer.SetFd(fd)
	if err = w.initFd(); err != nil {
		return err
	}
	return nil
}

// Init file logger with json config.
// config like:
//	{
//	"filename":"log/gogs.log",
//	"maxlines":10000,
//	"maxsize":1<<30,
//	"daily":true,
//	"maxdays":15,
//	"rotate":true
//	}
func (w *file) Init(config string) error {
	if err := json.Unmarshal([]byte(config), w); err != nil {
		return err
	}
	if len(w.Filename) == 0 {
		return errors.New("config must have filename")
	}
	return w.start()
}

func (w *file) docheck(size int) {
	w.startLock.Lock()
	defer w.startLock.Unlock()
	if w.Rotate && ((w.Maxlines > 0 && w.currentlines >= w.Maxlines) ||
		(w.Maxsize > 0 && w.currentSize >= w.Maxsize) ||
		(w.Daily && time.Now().Day() != w.opendate)) {
		if err := w.DoRotate(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.Filename, err)
			return
		}
	}
	w.currentlines++
	w.currentSize += size
}

// write logger message into file.
func (w *file) Write(msg string, skip int, level Level) error {
	if level < w.Level {
		return nil
	}
	n := 24 + len(msg) // 24 stand for the length "2013/06/23 21:00:22 [T] "
	w.docheck(n)
	w.Logger.Println(msg)
	return nil
}

func (w *file) createLogFile() (*os.File, error) {
	// Open the log file
	return os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
}

func (w *file) initFd() error {
	fd := w.writer.file
	finfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat: %s\n", err)
	}
	w.currentSize = int(finfo.Size())
	w.opendate = time.Now().Day()
	if finfo.Size() > 0 {
		content, err := ioutil.ReadFile(w.Filename)
		if err != nil {
			return err
		}
		w.currentlines = len(strings.Split(string(content), "\n"))
	} else {
		w.currentlines = 0
	}
	return nil
}

// DoRotate means it need to write file in new file.
// new file name like xx.log.2013-01-01.2
func (w *file) DoRotate() error {
	_, err := os.Lstat(w.Filename)
	if err == nil { // file exists
		// Find the next available number
		num := 1
		fname := ""
		for ; err == nil && num <= 999; num++ {
			fname = w.Filename + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), num)
			_, err = os.Lstat(fname)
		}
		// return error if the last file checked still existed
		if err == nil {
			return fmt.Errorf("rotate: cannot find free log number to rename %s\n", w.Filename)
		}

		// block Logger's io.Writer
		w.writer.Lock()
		defer w.writer.Unlock()

		fd := w.writer.file
		fd.Close()

		// close fd before rename
		// Rename the file to its newfound home
		if err = os.Rename(w.Filename, fname); err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}

		// re-start logger
		if err = w.start(); err != nil {
			return fmt.Errorf("Rotate StartLogger: %s\n", err)
		}

		go w.deleteOldLog()
	}

	return nil
}

func (w *file) deleteOldLog() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("Unable to delete old log '%s', error: %+v", path, r)
			}
		}()

		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-60*60*24*w.Maxdays) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.Filename)) {
				os.Remove(path)
			}
		}
		return returnErr
	})
}

// destroy file logger, close file writer.
func (w *file) Destroy() error {
	return w.writer.file.Close()
}

// flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (w *file) Flush() {
	w.writer.file.Sync()
}

func init() {
	Register("file", newFile)
}
