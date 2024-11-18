package main

import (
	"flag"
	"fmt"
	"httpqueue/cmd/server/handler"
	"httpqueue/lib/middleware"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	LongPollingTimeout int
}

func loadConfig() Config {
	config := Config{}

	// Define flags
	flag.StringVar(&config.Port, "port", "8080", "Port to run the server on")
	flag.IntVar(&config.LongPollingTimeout, "long-polling-timeout", 10, "Long Polling Timeout")
	flag.Parse()

	// Overwrite with envvars if they exist
	if envPort := os.Getenv("PORT"); envPort != "" {
		config.Port = envPort
	}
	if envLPT := os.Getenv("LONG_POLLING_TIMEOUT"); envLPT != "" {
		lpt, err := strconv.Atoi(envLPT)
		if err != nil {
			log.Printf("Error parsing LONG_POLLING_TIMEOUT: %v", err)
		} else {
			config.LongPollingTimeout = lpt
		}
	}

	return config
}

func main() {
	config := loadConfig()

	handler := handler.NewQueueHandler(config.LongPollingTimeout)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /queues/{queueName}", handler.HandlePush)
	mux.HandleFunc("GET /queues/{queueName}", handler.HandlePull)

	// Apply logging middleware
	loggedMux := middleware.LoggingMiddleware(mux)

	bindAddress := fmt.Sprintf(":%s", config.Port)

	fmt.Printf("Server starting on %s with long polling timeout %d seconds\n", bindAddress, config.LongPollingTimeout)
	if err := http.ListenAndServe(bindAddress, loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
