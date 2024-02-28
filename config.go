package main

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type config struct {
	port                 int
	originalRobotsURL    *url.URL
	timeoutRobotsRequest time.Duration
	additionalRobotsFile string
	newRobotsEndpoint    string
}

func loadConfigFromEnv() (cfg config, err error) {
	port := cmp.Or(os.Getenv("PORT"), "80")
	cfg.port, err = strconv.Atoi(port)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse port: %w", err)
	}

	robotsUrl := os.Getenv("ORIGINAL_ROBOTS_URL")
	if robotsUrl == "" {
		return cfg, fmt.Errorf("original robots URL (`ORIGINAL_ROBOTS_URL`) is required")
	}

	timeout := cmp.Or(os.Getenv("TIMEOUT_ROBOTS_REQUEST_SECONDS"), "5s")
	cfg.timeoutRobotsRequest, err = time.ParseDuration(timeout)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse timeout seconds: %w", err)
	}

	cfg.originalRobotsURL, err = url.Parse(robotsUrl)
	if err != nil {
		return cfg, fmt.Errorf("failed to get original robots URL: %w", err)
	}

	additionalRobotsFilePath := cmp.Or(os.Getenv("ADDITIONAL_ROBOTS_FILE"), "additional_robots.txt")
	additionalRobotsFile, err := os.Open(additionalRobotsFilePath)
	if err != nil {
		return cfg, fmt.Errorf("failed to open additional robots file: %w", err)
	}
	defer func() {
		// note: take care to extend this if additional errors can occur after this point
		if closeErr := additionalRobotsFile.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close additional robots file: %w", closeErr)
		}
	}()

	endpoint := cmp.Or(os.Getenv("ENDPOINT"), "robots.txt")
	cfg.newRobotsEndpoint = endpoint

	return cfg, nil
}
