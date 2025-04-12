package clearlog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"unicode/utf8"
)

type Logger struct {
	out         io.WriteCloser
	shouldClose bool

	// LogLevel is the minimum level of logging.
	LogLevel LogLevel
	// Err is the last error encountered during logging
	Err error
	// OnError indicates what to do if there's an error during logging.
	OnError OnError
}

// KV is a key-value pair.
type KV struct {
	Key   string
	Value any
}

// LogLevel is the level of logging.
type LogLevel int

const (
	DEBUG LogLevel = 10
	INFO           = 20
	WARN           = 30
	ERROR          = 40
	FATAL          = 50
)

// OnError ...
type OnError int

const (
	Ignore OnError = iota
	Stop
	Panic
)

// NewStderrLogger creates a new logger to stderr.
func NewStderrLogger() *Logger {
	return &Logger{
		out:         os.Stderr,
		shouldClose: false,
		LogLevel:    DEBUG,
	}
}

// NewFileLogger creates a new logger to the specified file. The file is truncated if it exists.
func NewFileLogger(path string) (*Logger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &Logger{
		out:         f,
		shouldClose: true,
		LogLevel:    DEBUG,
	}, nil
}

// NewLogger creates a new logger with the given output stream.
func NewLogger(out io.WriteCloser) *Logger {
	return &Logger{
		out:         out,
		shouldClose: false,
		LogLevel:    DEBUG,
	}
}

func (l *Logger) Close() error {
	var err error
	if l.shouldClose {
		err = l.out.Close()
	}
	l.out = nil
	return err
}

func (l *Logger) Debug(kvs ...KV) {
	l.log(DEBUG, kvs...)
}

func (l *Logger) Info(kvs ...KV) {
	l.log(INFO, kvs...)
}

func (l *Logger) Warning(kvs ...KV) {
	l.log(WARN, kvs...)
}

func (l *Logger) Error(kvs ...KV) {
	l.log(ERROR, kvs...)
}

func (l *Logger) Fatal(kvs ...KV) {
	l.log(FATAL, kvs...)
}

func (l *Logger) Log(level LogLevel, kvs ...KV) {
	l.log(level, kvs...)
}

func (l *Logger) log(level LogLevel, kvs ...KV) {
	if l.out == nil {
		return
	}
	if level < l.LogLevel {
		// Filter logs below the minimum level.
		return
	}
	l.logKeyValue("level", level)
	l.printf(" ")
	file, line, pkg, funcName := getContext(2)
	l.logKeyValue("file", file)
	l.printf(" ")
	l.logKeyValue("line", line)
	l.printf(" ")
	l.logKeyValue("package", pkg)
	l.printf(" ")
	l.logKeyValue("func", funcName)
	for _, kv := range kvs {
		l.printf(" ")
		l.logKeyValue(kv.Key, kv.Value)
	}
	l.printf("\n")
}

var (
	funcNameRE = regexp.MustCompile(`^(.*)\.(\w+)$`)
)

func getContext(skip int) (string, int, string, string) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		file = "unknown file"
		line = 0
	}
	file = filepath.Base(file)
	fn := runtime.FuncForPC(pc)
	pkg, funcName := "unknown_package", "unknown_function"
	if fn != nil {
		parts := funcNameRE.FindStringSubmatch(fn.Name())
		pkg, funcName = parts[1], parts[2]
	}
	return file, line, pkg, funcName
}

func (l *Logger) logKeyValue(key string, value any) {
	l.escape(key)
	l.printf("=")
	l.escape(toString(value))
}

func (l *Logger) printf(format string, args ...any) {
	_, err := fmt.Fprintf(l.out, format, args...)
	if err == nil {
		return
	}
	l.Err = err
	switch l.OnError {
	case Ignore:
		// Do nothing
	case Stop:
		// Close the logger, preserving any eventual error during close.
		l.Err = errors.Join(err, l.Close())
	case Panic:
		panic(err)
	}
}

var (
	symbolRE = regexp.MustCompile(`[^\pN\pL_]`)
)

func (l *Logger) escape(text string) {
	var hasReplacement bool
	replaced := symbolRE.ReplaceAllStringFunc(text, func(ch string) string {
		hasReplacement = true
		r, _ := utf8.DecodeRuneInString(ch)
		if r == '"' {
			return `\"`
		}
		if r == '\\' {
			return `\\`
		}
		if r == '\n' {
			return `\\n`
		}
		return ch
	})
	if !hasReplacement {
		l.printf(text)
		return
	}
	l.printf(`"`)
	l.printf(replaced)
	l.printf(`"`)
}

func toString(x any) string {
	return fmt.Sprintf("%v", x)
}
