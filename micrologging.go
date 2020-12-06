package micrologging

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the loglevel, from TRACE (0) to FATAL (6)
type Level uint8

const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

const defaultLogLevel = INFO

const timeFormat = "2006-01-02 15:04:05.000"

func (l Level) String() string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "?????"
	}
}

// RootLogger is the entry point to logging. You can get either the root looger, or
// construct your own Logger, that will be a child of the root logger.
type rootLogger struct {
	mu      sync.Mutex
	outputs []io.Writer
	level   Level
}

var root *rootLogger

type Logger struct {
	name string
}

func init() {

	outputs := make([]io.Writer, 1)
	outputs[0] = os.Stdout

	root = &rootLogger{
		outputs: outputs,
		level:   defaultLogLevel,
	}
}

//LevelFromString reconstructs the loglevel from a given
//string. It is case insensitive and recognizes some
//commong loglevel name synonyms, such as WARN=WARNING,
//ERR=ERROR
func LevelFromString(str string) (Level, error) {
	str = strings.ToUpper(str)

	switch str {
	case "TRACE":
		return TRACE, nil
	case "DEBUG":
		return DEBUG, nil
	case "INFO":
		return INFO, nil
	case "WARN":
		fallthrough
	case "WARNING":
		return WARN, nil
	case "ERR":
		fallthrough
	case "ERROR":
		return ERROR, nil
	case "FATAL":
		return FATAL, nil
	default:
		return defaultLogLevel, fmt.Errorf("No such log level %s", str)
	}
}

//SetRootOutput assigns an io.Writer to root logger
func AddRootOutput(output io.Writer) {
	root.mu.Lock()
	defer root.mu.Unlock()
	root.outputs = append(root.outputs, output)
}

//SetRootLevel sets the loglvevel of the root logger
func SetRootLevel(level Level) {
	root.level = level
}

//GetLogger constructs the child logger of the root with specified name
func GetLogger(name string) *Logger {
	return &Logger{
		name,
	}
}

//Printf logs the line with given loglevel, formatted according to format, using
//the root logger
func (l *Logger) Printf(level Level, format string, messages ...interface{}) {
	root.printf(level, format, l.name, messages...)
}

func (l *rootLogger) printf(level Level, format, name string, messages ...interface{}) {

	if level < l.level {
		return
	}

	b := &strings.Builder{}

	b.WriteString("(" + time.Now().Format(timeFormat) + ") ")

	b.WriteString("[")
	b.WriteString(level.String())
	b.WriteString("] ")

	if name != "" {
		b.WriteString("(")
		b.WriteString(name)
		b.WriteString(") ")
	}

	if len(messages) > 0 {
		b.WriteString(fmt.Sprintf(format, messages...))
	} else {
		b.WriteString(format)
	}

	out := strings.TrimSpace(b.String())

	for _, output := range l.outputs {
		if syslogWriter, ok := output.(*syslog.Writer); ok {

			out := fmt.Sprintf(format, messages...)

			switch level {
			case TRACE:
				fallthrough
			case DEBUG:
				syslogWriter.Debug(out)
			case INFO:
				syslogWriter.Info(out)
			case WARN:
				syslogWriter.Warning(out)
			case ERROR:
				syslogWriter.Err(out)
			case FATAL:
				syslogWriter.Crit(out)
			default:
				syslogWriter.Info(out)
			}
			continue
		}

		fmt.Fprintln(output, out)
	}

}

func (l *Logger) Trace(format string, messages ...interface{}) {
	l.Printf(TRACE, format, messages...)
}

func (l *Logger) Debug(format string, messages ...interface{}) {
	l.Printf(DEBUG, format, messages...)
}

func (l *Logger) Info(format string, messages ...interface{}) {
	l.Printf(INFO, format, messages...)
}

func (l *Logger) Warn(format string, messages ...interface{}) {
	l.Printf(WARN, format, messages...)
}

func (l *Logger) Error(format string, messages ...interface{}) {
	l.Printf(ERROR, format, messages...)
}

func (l *Logger) Fatal(format string, messages ...interface{}) {
	l.Printf(FATAL, format, messages...)
}

func Trace(format string, messages ...interface{}) {
	root.printf(TRACE, format, "", messages...)
}

func Debug(format string, messages ...interface{}) {
	root.printf(DEBUG, format, "", messages...)
}

func Info(format string, messages ...interface{}) {
	root.printf(INFO, format, "", messages...)
}

func Warn(format string, messages ...interface{}) {
	root.printf(WARN, format, "", messages...)
}

func Error(format string, messages ...interface{}) {
	root.printf(ERROR, format, "", messages...)
}

func Fatal(format string, messages ...interface{}) {
	root.printf(FATAL, format, "", messages...)
}
