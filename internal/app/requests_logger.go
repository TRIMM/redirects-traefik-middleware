package app

import (
	"bufio"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"redneck-traefik-middleware/api/v1"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	fileName    string
	requestsMap map[string]time.Time
	mutex       sync.Mutex
	gqlClient   *v1.GraphQLClient
}

func NewLogger(fileName string, gqlClient *v1.GraphQLClient) *Logger {
	return &Logger{
		requestsMap: make(map[string]time.Time),
		mutex:       sync.Mutex{},
		fileName:    fileName,
		gqlClient:   gqlClient,
	}
}

// SendLogsWeekly starts a cron job to send request logs weekly
func (l *Logger) SendLogsWeekly() {
	c := cron.New()
	_, err := c.AddFunc("@weekly", func() {
		l.SendLogs()
	})
	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}
	c.Start()
}

func (l *Logger) SendLogs() {
	err := l.LoadLoggedRequests()
	if err != nil {
		log.Println(err)
	}

	response, err := l.gqlClient.ExecuteLogRequestsMutation(&l.requestsMap)
	if err != nil {
		log.Println("Failed to execute GraphQL mutation:", err)
	}

	fmt.Println(response.Message)
}

// LoadLoggedRequests loads the requestsMap with data from the .log file
func (l *Logger) LoadLoggedRequests() error {
	file, err := os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}()

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

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}()

	// Write to log file
	logEntry := fmt.Sprintf("%s,%s\n", requestURL, time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}
