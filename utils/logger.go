package utils

import (
	"io"
	"log"
	"os"

	logging "github.com/op/go-logging"
)

const (
	defaultLogFormat = "[%{module}:%{level}] %{message}"
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

	l.initLevel().
		setFormat().
		setLevels().
		setBackend()

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

func (l *Logger) setLevels() *Logger {
	for _, s := range l.config.Modules {
		logging.SetLevel(logging.Level(l.config.level), s)
	}
	return l
}

func (l *Logger) setFormat() *Logger {
	format := logging.MustStringFormatter(l.config.Format)
	logging.SetFormatter(format)
	return l
}

func (l *Logger) setBackend() *Logger {
	switch l.config.Backend {
	case "os.stdout":
		l.config.Output = os.Stdout
	default:
		l.config.Output = os.Stdout
	}

	backend := logging.NewLogBackend(l.config.Output, "", 0)
	backend.Color = l.config.Colored
	logging.SetBackend(backend)

	return l
}

func (l *Logger) GetModuleLogger(module string) *Logger {
	l.Logger = logging.MustGetLogger(module)

	return l
}
