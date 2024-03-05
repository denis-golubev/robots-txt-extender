package main

import (
	"cmp"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"time"
	"unicode/utf8"
)

// config contains the configuration for the service.
// Please refer to the README for more details
// about the configuration.
type config struct {
	port                   int
	originalRobotsURL      *url.URL
	timeoutRobotsRequest   time.Duration
	additionalRobotsFile   string
	newRobotsEndpoint      string
	includeOriginalHeaders bool
	logLevel               slog.Level
	xForwardedProto        *string
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
	cfg.originalRobotsURL, err = url.Parse(robotsUrl)
	if err != nil {
		return cfg, fmt.Errorf("failed to get original robots URL: %w", err)
	}

	timeout := cmp.Or(os.Getenv("TIMEOUT_ROBOTS_REQUEST_SECONDS"), "5s")
	cfg.timeoutRobotsRequest, err = time.ParseDuration(timeout)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse timeout seconds: %w", err)
	}

	additionalRobotsFilePath := cmp.Or(os.Getenv("ADDITIONAL_ROBOTS_FILE"), "additional_robots.txt")
	additionalRobotsFile, err := os.ReadFile(additionalRobotsFilePath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read additional robots file: %w", err)
	}

	if !utf8.Valid(additionalRobotsFile) {
		return cfg, fmt.Errorf("additional robots file is not valid UTF-8")
	}
	cfg.additionalRobotsFile = string(additionalRobotsFile)

	endpoint := cmp.Or(os.Getenv("ENDPOINT"), "robots.txt")
	cfg.newRobotsEndpoint = endpoint

	includeOriginalHeaders := cmp.Or(os.Getenv("INCLUDE_ORIGINAL_HEADERS"), "true")
	cfg.includeOriginalHeaders, err = strconv.ParseBool(includeOriginalHeaders)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse \"include original headers\" setting: %w", err)
	}

	logLevel := cmp.Or(os.Getenv("LOG_LEVEL"), "info")
	switch logLevel {
	case "debug":
		cfg.logLevel = slog.LevelDebug
	case "info":
		cfg.logLevel = slog.LevelInfo
	case "warn":
		cfg.logLevel = slog.LevelWarn
	case "error":
		cfg.logLevel = slog.LevelError
	default:
		return cfg, fmt.Errorf("unknown log level: %s", logLevel)
	}

	xForwardedProto := os.Getenv("X_FORWARDED_PROTO")
	if xForwardedProto != "" {
		cfg.xForwardedProto = &xForwardedProto
	}

	return cfg, nil
}
