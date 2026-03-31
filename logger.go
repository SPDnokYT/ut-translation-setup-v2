package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileLogger struct {
	filePath string
	file     *os.File
}

func NewFileLogger(baseName string) (*FileLogger, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error obtaining executable path: %v", err)
	}

	timestamp := time.Now().Format(time.RFC3339)
	exeDir := filepath.Dir(execPath)
	filename := fmt.Sprintf("%s_%s.log", baseName, timestamp)
	logPath := filepath.Join(exeDir, filename)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening the log file: %v", err)
	}

	return &FileLogger{
		filePath: logPath,
		file:     file,
	}, nil
}

func (l *FileLogger) GetLogPath() string {
	return l.filePath
}

func (l *FileLogger) write(level, message string) {
	timestamp := time.Now().Format(time.RFC3339)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	l.file.WriteString(logLine)
	fmt.Print(logLine)
}

func (l *FileLogger) Print(message string)   { l.write("PRINT", message) }
func (l *FileLogger) Trace(message string)   { l.write("TRACE", message) }
func (l *FileLogger) Debug(message string)   { l.write("DEBUG", message) }
func (l *FileLogger) Info(message string)    { l.write("INFO", message) }
func (l *FileLogger) Warning(message string) { l.write("WARN", message) }
func (l *FileLogger) Error(message string)   { l.write("ERROR", message) }
func (l *FileLogger) Fatal(message string) {
	l.write("FATAL", message)
	os.Exit(1)
}
