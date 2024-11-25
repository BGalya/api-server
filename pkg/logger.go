package api_sec

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func CreateLogFile() error {
	// Create the log file (will create a new file or truncate if it exists)
	file, err := os.Create("access.log")
	if err != nil {
		return fmt.Errorf("error creating log file: %v", err)
	}
	defer file.Close()
	return nil
}

// LogEntry represents the structure for the log entry (request and response details)
type LogEntry struct {
	Req struct {
		URL        string            `json:"url"`
		QSParams   string            `json:"qs_params"`
		Headers    map[string]string `json:"headers"`
		ReqBodyLen int               `json:"req_body_len"`
	} `json:"req"`
	Rsp struct {
		StatusClass string `json:"status_class"`
		RspBodyLen  int    `json:"rsp_body_len"`
	} `json:"rsp"`
}

// WriteToFile writes the request and response details to the log file in JSON format
func WriteToFile(r *http.Request, statusCode int, rspBodyLen int) {
	// Open the log file (or create if it doesn't exist)
	file, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening or creating log file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a LogEntry to hold the request and response details
	logEntry := LogEntry{}

	// Fill the request details
	logEntry.Req.URL = r.URL.String()
	logEntry.Req.QSParams = r.URL.RawQuery
	logEntry.Req.Headers = make(map[string]string)
	for key, values := range r.Header {
		logEntry.Req.Headers[key] = fmt.Sprintf("%s", values)
	}
	logEntry.Req.ReqBodyLen = int(r.ContentLength)

	// Determine the status class (2xx, 3xx, 4xx, 5xx)
	statusClass := getStatusClass(statusCode)
	logEntry.Rsp.StatusClass = statusClass
	logEntry.Rsp.RspBodyLen = rspBodyLen

	// Marshal the log entry to JSON
	logJSON, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("Error marshaling log entry: %v\n", err)
		return
	}

	// Write the JSON log entry to the file (single line)
	_, err = file.Write(append(logJSON, '\n'))
	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

// getStatusClass determines the status class (2xx, 3xx, 4xx, 5xx)
func getStatusClass(statusCode int) string {
	switch statusCode / 100 {
	case 2:
		return "2xx"
	case 3:
		return "3xx"
	case 4:
		return "4xx"
	case 5:
		return "5xx"
	default:
		return "unknown"
	}
}
