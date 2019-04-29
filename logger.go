package spm

import (
	"bufio"
	"fmt"

	"math/rand"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Prefix   []byte
	LogColor int

	Logfile  *lumberjack.Logger
	filename string
}

func NewLogging(name string) (*Logger, error) {
	linkname := "/tmp/spm/" + name + ".log"
	logfile := &lumberjack.Logger{
		Filename:   linkname,
		MaxSize:    1024, //mb
		MaxBackups: 10,
		MaxAge:     7, // days
		Compress:   false,
	}

	// create random log color code
	code := genColorCode()

	prefix := loggerPrefix(code, name)

	return &Logger{
		Prefix:   []byte(prefix),
		LogColor: code,
		Logfile:  logfile,
		filename: linkname,
	}, nil
}

func (l *Logger) FileName() string {
	return l.filename
}

// Write writes given string into Logfile
func (l *Logger) Write(s []byte) error {
	if _, err := l.Logfile.Write(s); err != nil {
		return err
	}
	return nil
}

var ln = []byte("\n")

// Output reads the in then writes into both stdout and logfile
func (l *Logger) Output(in *bufio.Scanner) error {
	for in.Scan() {
		_ = l.Write(l.Prefix)
		_ = l.Write(in.Bytes())
		_ = l.Write(ln)
	}

	if err := in.Err(); err != nil {
		return err
	}

	return nil
}

func (l *Logger) Close() error {
	if err := l.Logfile.Close(); err != nil {
		return err
	}

	return nil
}

// LoggerPrefix wraps given string and time with unix color code, as prefix
func loggerPrefix(code int, s string) string {
	t := time.Now().Format("15:04:05 PM")
	if isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Sprintf("\033[38;5;%dm%s %s | \033[0m", code, t, s)
	}
	return fmt.Sprintf("%s %s | ", t, s)
}

func genColorCode() (code int) {
	rand.Seed(int64(time.Now().Nanosecond()))
	code = rand.Intn(231) + 1
	return
}
