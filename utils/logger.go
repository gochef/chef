package utils

import (
	"io"
	"log"
	"os"

	logging "github.com/op/go-logging"
)

const (
	defaultLogFormat = "[%{module}.%{shortfunc}.%{level} %{time:15:04:05}] %{message}"
)

// Log levels.
const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"NOTICE",
	"INFO",
	"DEBUG",
}

type (
	// Level embedes the logging's level
	Level int

	// LoggerConfig sets logger configuration
	LoggerConfig struct {
		level Level
		// Level convertes to level during initialization
		Level string
		// format
		Format string
		// enable colors
		Colored bool
		Backend string
		File    string
		Modules []string
		Output  io.Writer
	}

	// Logger represents a logger intance
	Logger struct {
		config *LoggerConfig
		*logging.Logger
	}
)

// NewLogger returns a logger instance
func NewLogger(config *LoggerConfig) *Logger {
	l := &Logger{
		config: config,
	}

	l.initLevel().setBackends()

	return l
}

func (l *Logger) initLevel() *Logger {
	level, err := logging.LogLevel(l.config.Level)
	if err != nil {
		log.Panicf("Invalid log level %s: %v", l.config.Level, err)
	}
	l.config.level = Level(level)
	return l
}

func (l *Logger) setLevel(b logging.LeveledBackend) {
	for _, s := range l.config.Modules {
		b.SetLevel(logging.Level(l.config.level), s)
	}
}

func (l *Logger) setBackends() *Logger {
	format := logging.MustStringFormatter(l.config.Format)
	screenBackend := l.getScreenBackend(format)
	fileBackend := l.getFileBackend(format)

	logging.SetBackend(screenBackend, fileBackend)

	return l
}

func (l *Logger) getFileBackend(format logging.Formatter) logging.Backend {
	file, err := os.OpenFile(l.config.File, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	backendFile := logging.NewLogBackend(file, "", 0)
	backendFileFormatter := logging.NewBackendFormatter(backendFile, format)
	backendFile.Color = l.config.Colored
	backendFileLeveled := logging.AddModuleLevel(backendFileFormatter)
	return backendFileLeveled
}

func (l *Logger) getScreenBackend(format logging.Formatter) logging.LeveledBackend {
	backendScreen := logging.NewLogBackend(os.Stdout, "", 0)
	backendScreen.Color = l.config.Colored
	backendScreenFormatter := logging.NewBackendFormatter(backendScreen, format)
	backendScreenLeveled := logging.AddModuleLevel(backendScreenFormatter)
	return backendScreenLeveled
}

func (l *Logger) GetModuleLogger(module string) *Logger {
	l.Logger = logging.MustGetLogger(module)

	return l
}
