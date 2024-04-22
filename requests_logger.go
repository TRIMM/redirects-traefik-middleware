package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	fileName    string
	requestsMap map[string]time.Time
	mutex       sync.Mutex
	gqlClient   *GraphQLClient
}

func NewLogger(fileName string, gqlClient *GraphQLClient) *Logger {
	return &Logger{
		requestsMap: make(map[string]time.Time),
		mutex:       sync.Mutex{},
		fileName:    fileName,
		gqlClient:   gqlClient,
	}
}

func (l *Logger) SendLogs() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mutex.Lock()
			err := l.LoadLoggedRequests()
			if err != nil {
				log.Println(err)
			}

			response, err := l.gqlClient.executeLogRequestsMutation(&l.requestsMap)
			if err != nil {
				log.Println("Failed to execute GraphQL mutation:", err)
			}

			fmt.Println(response.Message)
			l.mutex.Unlock()
		}
	}

}

// LoadLoggedRequests loads the requestsMap with data from the .log file
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

		if existingTimestamp, ok := l.requestsMap[requestURL]; !ok || timestamp.After(existingTimestamp) {
			l.requestsMap[requestURL] = timestamp
		}
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
