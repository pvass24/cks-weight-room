package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents the severity of a log message
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// Logger is the global logger instance
type Logger struct {
	level      Level
	file       *os.File
	stdLogger  *log.Logger
	mu         sync.Mutex
	logDir     string
	maxSize    int64 // Maximum size in bytes before rotation
	maxFiles   int   // Maximum number of rotated files to keep
	currentSize int64
}

var (
	globalLogger *Logger
	once         sync.Once
)

// Config holds logger configuration
type Config struct {
	LogDir   string
	Level    Level
	MaxSize  int64 // In MB
	MaxFiles int
}

// Init initializes the global logger
func Init(config Config) error {
	var err error
	once.Do(func() {
		globalLogger, err = newLogger(config)
	})
	return err
}

// newLogger creates a new logger instance
func newLogger(config Config) (*Logger, error) {
	// Default values
	if config.LogDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.LogDir = filepath.Join(homeDir, ".cks-weight-room", "logs")
	}
	if config.MaxSize == 0 {
		config.MaxSize = 10 // 10MB default
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 5
	}

	// Create log directory
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		level:    config.Level,
		logDir:   config.LogDir,
		maxSize:  config.MaxSize * 1024 * 1024, // Convert MB to bytes
		maxFiles: config.MaxFiles,
	}

	// Open log file
	if err := logger.openLogFile(); err != nil {
		return nil, err
	}

	return logger, nil
}

// openLogFile opens the current log file
func (l *Logger) openLogFile() error {
	logPath := filepath.Join(l.logDir, "cks-weight-room.log")

	// Get file size if exists
	if info, err := os.Stat(logPath); err == nil {
		l.currentSize = info.Size()
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file

	// Create multi-writer for both file and stdout (in debug mode)
	var writers []io.Writer
	writers = append(writers, file)
	if l.level == LevelDebug {
		writers = append(writers, os.Stdout)
	}

	l.stdLogger = log.New(io.MultiWriter(writers...), "", 0)
	return nil
}

// rotate rotates the log file
func (l *Logger) rotate() error {
	// Close current file
	if l.file != nil {
		l.file.Close()
	}

	// Rotate existing files
	logPath := filepath.Join(l.logDir, "cks-weight-room.log")
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := filepath.Join(l.logDir, fmt.Sprintf("cks-weight-room-%s.log", timestamp))

	if err := os.Rename(logPath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Clean up old log files
	l.cleanupOldLogs()

	// Reset size counter
	l.currentSize = 0

	// Open new log file
	return l.openLogFile()
}

// cleanupOldLogs removes old rotated log files
func (l *Logger) cleanupOldLogs() {
	files, err := filepath.Glob(filepath.Join(l.logDir, "cks-weight-room-*.log"))
	if err != nil || len(files) <= l.maxFiles {
		return
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var fileInfos []fileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: f, modTime: info.ModTime()})
	}

	// Remove oldest files
	numToRemove := len(fileInfos) - l.maxFiles
	for i := 0; i < numToRemove && i < len(fileInfos); i++ {
		os.Remove(fileInfos[i].path)
	}
}

// log writes a log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if rotation is needed
	if l.currentSize >= l.maxSize {
		if err := l.rotate(); err != nil {
			// If rotation fails, log to stderr
			fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
		}
	}

	// Format message
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := levelNames[level]
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, levelStr, message)

	// Write to log
	n, err := l.stdLogger.Writer().Write([]byte(logLine))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
		return
	}
	l.currentSize += int64(n)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Close closes the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Global logging functions

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

// Close closes the global logger
func Close() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

// GetLogDir returns the log directory path
func GetLogDir() string {
	if globalLogger != nil {
		return globalLogger.logDir
	}
	return ""
}
