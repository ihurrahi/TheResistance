package logger

import (
    "path/filepath"
    "os"
    "fmt"
)

const (
    logPath = "logs"
)

type Logger struct {
    path string
    file *os.File
}

var loggers []*Logger = make([]*Logger, 0)

func (logger *Logger) LogMessage(message string) {
    if logger.file != nil {
        logger.file.Sync()
        logger.file.WriteString(message)
        logger.file.WriteString("\n")
    }
}

func InitializeLogger(logFile string) *Logger {
    var logger *Logger
    var fullPath string = filepath.Join(logPath, logFile)
    for i := 0; i < len(loggers); i++ {
        if loggers[i].path == fullPath {
            logger = loggers[i]
        }
    }
    if logger == nil {
        logger = new(Logger)
        logger.path = fullPath
        file, err := os.OpenFile(fullPath, os.O_RDWR, 0666)
        if err != nil {
            file, err = os.Create(fullPath)
            if err != nil {
                fmt.Fprintln(os.Stdout, "Error creating file " + fullPath + ". Won't be able to log to it.")
            }
        }
        logger.file = file
        loggers = append(loggers, logger)
    }
    return logger
}