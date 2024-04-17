package logr

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	DefaultCallerDepth = 3
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	Logrer             *logr
	dayFormat          = "20060102"
	setuponce          sync.Once
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

type PrintfFunc func(l *logr, format string, v ...interface{})

type Config struct {
	FilePath       string
	Encoding       string
	FileNamePrefix string
	PrintfFunc     PrintfFunc
}

type logr struct {
	config  Config
	logFile *os.File
	logger  *log.Logger
}

func Setup(lc Config) error {
	var err error

	setuponce.Do(func() {
		if lc.FilePath == "" {
			lc.FilePath = "runtime/logs"
		}
		if lc.FileNamePrefix == "" {
			lc.FileNamePrefix = "log"
		}
		if lc.Encoding == "" {
			lc.Encoding = "plain"
		}

		Logrer = &logr{config: lc}

		if f, e := MustOpen(Logrer.getLogFileName(), lc.FilePath); e == nil {
			Logrer.logFile = f
		} else {
			err = e
			return
		}

		Logrer.logger = log.New(Logrer.logFile, "", log.LstdFlags)
	})

	return err
}

func (l *logr) setLogFile() {
	temp := l.getLogFileName()
	if l.config.FilePath+"/"+temp != l.logFile.Name() {
		// 关闭之前的文件
		l.logFile.Close()

		// 重新打开文件
		if f, err := MustOpen(temp, l.config.FilePath); err != nil {
			log.Println(err)
			return
		} else {
			l.logFile = f
		}

		l.logger.SetOutput(l.logFile)
	}
}

func (l *logr) getLogFileName() string {
	return fmt.Sprintf("%s-%s.log", l.config.FileNamePrefix, time.Now().Format(dayFormat))
}

func (l *logr) setLinePrefix(level Level) {
	_, f, line, ok := runtime.Caller(DefaultCallerDepth)
	var linePrefix string
	if ok {
		linePrefix = fmt.Sprintf("%s|%s:%d|", levelFlags[level], filepath.Base(f), line)
	} else {
		linePrefix = fmt.Sprintf("%s|", levelFlags[level])
	}

	l.logger.SetPrefix(linePrefix)
}

func (l *logr) GetLogger() *log.Logger {
	return l.logger
}

func (l *logr) Printf(format string, v ...interface{}) {
	if l.config.PrintfFunc == nil {
		l.logger.Printf(format+"\n", v...)
	} else {
		l.config.PrintfFunc(l, format, v...)
	}
}

func SetPrintfFunc(f PrintfFunc) {
	if Logrer == nil {
		Setup(Config{PrintfFunc: f})
	} else {
		Logrer.config.PrintfFunc = f
	}
}

func write(level Level, v ...interface{}) {
	if Logrer == nil {
		Setup(Config{})
	}

	Logrer.setLinePrefix(level)
	Logrer.setLogFile()

	switch level {
	case DEBUG, INFO, WARN, ERROR:
		Logrer.logger.Println(v...)
	case FATAL:
		Logrer.logger.Fatalln(v...)
	default:
		Logrer.logger.Println(v...)
	}
}

func writef(level Level, format string, v ...interface{}) {
	if Logrer == nil {
		Setup(Config{})
	}

	Logrer.setLinePrefix(level)
	Logrer.setLogFile()

	switch level {
	case DEBUG, INFO, WARN, ERROR:
		Logrer.logger.Printf(format+"\n", v...)
	case FATAL:
		Logrer.logger.Fatalf(format+"\n", v...)
	default:
		Logrer.logger.Printf(format+"\n", v...)
	}
}

func Debug(v ...interface{}) {
	write(DEBUG, v...)
}

func Debugf(format string, v ...interface{}) {
	writef(DEBUG, format, v...)
}

func Info(v ...interface{}) {
	write(INFO, v...)
}

func Infof(format string, v ...interface{}) {
	writef(INFO, format, v...)
}

func Warn(v ...interface{}) {
	write(WARN, v...)
}

func Warnf(format string, v ...interface{}) {
	writef(WARN, format, v...)
}

func Error(v ...interface{}) {
	write(ERROR, v...)
}

func Errorf(format string, v ...interface{}) {
	writef(ERROR, format, v...)
}

func Fatal(v ...interface{}) {
	write(FATAL, v...)
}

func Fatalf(format string, v ...interface{}) {
	writef(FATAL, format, v...)
}

func CheckNotExist(src string) bool {
	_, err := os.Stat(src)

	return os.IsNotExist(err)
}

func CheckPermission(src string) bool {
	_, err := os.Stat(src)

	return os.IsPermission(err)
}

func IsNotExistMkDir(src string) error {
	if notExist := CheckNotExist(src); notExist == true {
		if err := MkDir(src); err != nil {
			return err
		}
	}

	return nil
}

func MkDir(src string) error {
	err := os.MkdirAll(src, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func MustOpen(fileName, filePath string) (*os.File, error) {
	perm := CheckPermission(filePath)
	if perm == true {
		return nil, fmt.Errorf("permission denied: %s", filePath)
	}

	err := IsNotExistMkDir(filePath)
	if err != nil {
		return nil, fmt.Errorf("mkdir err: %s, %v", filePath, err)
	}
	f, err := Open(filePath+"/"+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file: %v", err)
	}

	return f, nil
}
