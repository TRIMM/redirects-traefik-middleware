package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	fileName   string
	requestMap map[string]time.Time
	mutex      sync.Mutex
}

func NewLogger(fileName string) *Logger {
	return &Logger{
		requestMap: make(map[string]time.Time),
		mutex:      sync.Mutex{},
		fileName:   fileName,
	}
}

// LoadLoggedRequests loads the requestMap with data from the .log file
func (l *Logger) LoadLoggedRequests() error {
	file, err := os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			continue
		}

		requestURL := fields[0]
		timestampString := fields[1]

		timestamp, err := time.Parse(time.RFC3339, timestampString)
		if err != nil {
			continue
		}

		l.mutex.Lock()
		if existingTimestamp, ok := l.requestMap[requestURL]; !ok || timestamp.After(existingTimestamp) {
			l.requestMap[requestURL] = timestamp
		}
		l.mutex.Unlock()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %v", err)
	}

	return nil
}

// LogRequest logs any incoming requests with dates to the .log file
func (l *Logger) LogRequest(requestURL string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	file, err := os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	// Write to log file
	logEntry := fmt.Sprintf("%s,%s\n", requestURL, time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}
