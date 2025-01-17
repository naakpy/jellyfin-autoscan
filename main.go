package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
)

var log = logrus.New()

func initLogger(level string) {
	file, err := os.OpenFile("jellyfin-autoscan.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open log file: ", err)
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		PadLevelText:     true,
	})
	
	switch level {
	case "DEBUG":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

func getRefreshLibraryTaskID(baseURL, apiKey string) (string, error) {
	tasksURL := fmt.Sprintf("%s/ScheduledTasks", baseURL)
	client := &http.Client{}

	req, err := http.NewRequest("GET", tasksURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "MediaBrowser Token="+apiKey)
	log.WithFields(logrus.Fields{
		"url":     tasksURL,
		"headers": req.Header,
	}).Debug("Sending GET request")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("received status code %d: %s", resp.StatusCode, string(body))
	}

	var tasks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return "", fmt.Errorf("error decoding JSON: %w", err)
	}

	for _, task := range tasks {
		if key, ok := task["Key"].(string); ok && key == "RefreshLibrary" {
			if id, ok := task["Id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("RefreshLibrary task not found")
}

func startTask(baseURL, apiKey, taskID string) error {
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	
	taskURL := fmt.Sprintf("%s/ScheduledTasks/Running/%s", baseURL, taskID)
	client := &http.Client{}

	req, err := http.NewRequest("POST", taskURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "MediaBrowser Token="+apiKey)
	log.WithFields(logrus.Fields{
		"url":     taskURL,
		"headers": req.Header,
	}).Debug("Sending POST request")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	log.WithField("statusCode", resp.StatusCode).Debug("Received response")
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status code 204, got %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func startServer(baseURL, apiKey string) error {
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		log.WithField("remoteAddr", r.RemoteAddr).Info("Received refresh request")
		
		if r.Method != http.MethodGet {
			log.WithField("method", r.Method).Info("Method not allowed")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Debug("Fetching RefreshLibrary task ID...")
		taskID, err := getRefreshLibraryTaskID(baseURL, apiKey)
		if err != nil {
			log.WithError(err).Error("Error getting task ID")
			http.Error(w, "Error getting task ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.WithField("taskID", taskID).Debug("Retrieved task ID")

		log.WithField("taskID", taskID).Debug("Starting refresh task")
		if err := startTask(baseURL, apiKey, taskID); err != nil {
			log.WithError(err).Error("Error starting task")
			http.Error(w, "Error starting task: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info("RefreshLibrary task started successfully")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "RefreshLibrary task started successfully")
	})

	log.Info("Server starting on :8282")
	return http.ListenAndServe(":8282", nil)
}

func loadConfig() (string, string, string, error) {
	baseURL := os.Getenv("JELLYFIN_BASE_URL")
	apiKey := os.Getenv("JELLYFIN_API_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

	if baseURL == "" || apiKey == "" {
		return "", "", "", fmt.Errorf("JELLYFIN_BASE_URL and JELLYFIN_API_KEY must be set in environment")
	}

	return baseURL, apiKey, logLevel, nil
}


func main() {
	baseURL, apiKey, logLevel, err := loadConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
		return
	}
	
	initLogger(logLevel)

	if err := startServer(baseURL, apiKey); err != nil {
		log.WithError(err).Error("Server failed to start")
	}
}