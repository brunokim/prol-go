package kif

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

type Logger struct {
	out         io.WriteCloser
	shouldClose bool
	encoder     logEncoder

	// LogLevel is the minimum level of logging.
	LogLevel LogLevel
	// Err is the last error encountered during logging
	Err error
	// OnError indicates what to do if there's an error during logging.
	OnError OnError
	// DisableCaller controls whether caller context is added to log messages.
	DisableCaller bool
}

// KV is a key-value pair.
type KV struct {
	Key   string
	Value any
}

// LogLevel is the level of logging.
type LogLevel int

//go:generate go run golang.org/x/tools/cmd/stringer -type LogLevel .
const (
	DEBUG LogLevel = 10
	INFO  LogLevel = 20
	WARN  LogLevel = 30
	ERROR LogLevel = 40
	FATAL LogLevel = 50
	PANIC LogLevel = 60
)

// OnError ...
type OnError int

const (
	Ignore OnError = iota
	Stop
	Panic
)

// Encoder
type Encoder int

const (
	LogfmtEncoder Encoder = iota
	JSONEncoder
)

// NewLogger creates a new logger writh the given output stream.
func NewLogger(out io.WriteCloser) *Logger {
	return &Logger{
		out:         out,
		shouldClose: false,
		encoder:     logfmtEncoder{},
		LogLevel:    DEBUG,
	}
}

// NewStderrLogger creates a new logger to stderr.
func NewStderrLogger() *Logger {
	return NewLogger(os.Stderr)
}

// NewFileLogger creates a new logger to the specified file. The file is truncated if it exists.
func NewFileLogger(path string) (*Logger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	logger := NewLogger(f)
	logger.shouldClose = true
	return logger, nil
}

func (l *Logger) Close() error {
	var err error
	if l.shouldClose {
		err = l.out.Close()
	}
	l.out = nil
	return err
}

func (l *Logger) SetEncoder(encoder Encoder) {
	switch encoder {
	case LogfmtEncoder:
		l.encoder = logfmtEncoder{}
	case JSONEncoder:
		l.encoder = jsonEncoder{}
	}
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
	if l == nil || l.out == nil {
		return
	}
	if level < l.LogLevel {
		// Filter logs below the minimum level.
		return
	}
	l.startLine()
	l.logEntry("level", level)
	if !l.DisableCaller {
		file, line, pkg, funcName := getContext(2)
		l.fieldSep()
		l.logEntry("file", file)
		l.fieldSep()
		l.logEntry("line", line)
		l.fieldSep()
		l.logEntry("package", pkg)
		l.fieldSep()
		l.logEntry("func", funcName)
	}
	for _, kv := range kvs {
		l.fieldSep()
		l.logEntry(kv.Key, kv.Value)
	}
	l.endLine()
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

func (l *Logger) logEntry(key string, value any) {
	l.writeString(l.encoder.escape(key))
	l.writeString(l.encoder.entrySep())
	l.writeString(l.encoder.toString(value))
}

func (l *Logger) writeString(text string) {
	_, err := l.out.Write([]byte(text))
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

func (l *Logger) startLine() {
	l.writeString(l.encoder.startLine())
}

func (l *Logger) fieldSep() {
	l.writeString(l.encoder.fieldSep())
}

func (l *Logger) endLine() {
	l.writeString(l.encoder.endLine())
}

type logEncoder interface {
	escape(text string) string
	toString(x any) string
	startLine() string
	fieldSep() string
	entrySep() string
	endLine() string
}

// ---

type logfmtEncoder struct{}

var (
	logfmtSymbolRE = regexp.MustCompile(`[\s'\\=]`)
)

func (logfmtEncoder) escape(text string) string {
	var hasReplacement bool
	replaced := logfmtSymbolRE.ReplaceAllStringFunc(text, func(ch string) string {
		hasReplacement = true
		r, _ := utf8.DecodeRuneInString(ch)
		if r == '\'' {
			return `\'`
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
		return text
	}
	return "'" + replaced + "'"
}

func (enc logfmtEncoder) toString(x any) string {
	return enc.escape(fmt.Sprintf("%v", x))
}

func (logfmtEncoder) startLine() string {
	return ""
}
func (logfmtEncoder) fieldSep() string {
	return " "
}
func (logfmtEncoder) entrySep() string {
	return "="
}
func (logfmtEncoder) endLine() string {
	return "\n"
}

// ---

type jsonEncoder struct{}

var (
	jsonControlChars = [...]string{
		`\u0000`, `\u0001`, `\u0002`, `\u0003`, `\u0004`, `\u0005`, `\u0006`, `\u0007`,
		`\b`, `\t`, `\n`, `\u000B`, `\u000C`, `\u000D`, `\u000E`, `\u000F`,
		`\u0010`, `\u0011`, `\f`, `\r`, `\u0014`, `\u0015`, `\u0016`, `\u0017`,
		`\u0018`, `\u0019`, `\u001A`, `\u001B`, `\u001C`, `\u001D`, `\u001E`, `\u001F`,
	}
)

func (jsonEncoder) escape(text string) string {
	var b strings.Builder
	b.WriteRune('"')
	for _, ch := range text {
		if ch < 0x20 {
			b.WriteString(jsonControlChars[ch])
		} else if ch == '"' {
			b.WriteString(`\"`)
		} else if ch == '\\' {
			b.WriteString(`\\`)
		} else {
			b.WriteRune(ch)
		}
	}
	b.WriteRune('"')
	return b.String()
}

func (enc jsonEncoder) toString(x any) string {
	return enc.escape(fmt.Sprintf("%v", x))
}

func (jsonEncoder) startLine() string {
	return "{"
}

func (jsonEncoder) fieldSep() string {
	return ","
}

func (jsonEncoder) entrySep() string {
	return ":"
}

func (jsonEncoder) endLine() string {
	return "}\n"
}
