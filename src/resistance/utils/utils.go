package utils

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

const (
	WSP_PORT            = "8081"
	GAME_REP_REQ_PORT   = "8082"
	GAME_PUB_SUB_PORT   = "8083"
	RESISTANCE_LOG_PATH = "logs/resistance.log"
	USER_LOG_PATH       = "logs/userLog.log"
	GAME_LOG_PATH       = "logs/gameLog.log"
)

// createLogger creates a logger that will log to the given file
func createLogger(filename string) (*log.Logger, *os.File, error) {
	logFile, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err.Error())
		logFile, err = os.Create(filename)
		if err != nil {
			return nil, logFile, errors.New("Error accessing access log file... Abort!")
		}
	}
	logger := log.New(logFile, "", log.LstdFlags)
	return logger, logFile, nil
}

// LogMessage logs the given message into the given log file.
func LogMessage(message string, logFileName string) {
	logger, logFile, err := createLogger(logFileName)
	defer logFile.Close()
	if err != nil {
		fmt.Println("Message not written to logs: " + message)
	}
	logger.Println(message)
}

// ConnectToDB connects to the local DB.
func ConnectToDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "resistance:resistance@unix(/var/run/mysql/mysql.sock)/resistance")

	if err != nil {
		LogMessage(err.Error(), RESISTANCE_LOG_PATH)
	}
	err = db.Ping()
	if err != nil {
		LogMessage(err.Error(), RESISTANCE_LOG_PATH)
		return nil, err
	}
	return db, nil
}
