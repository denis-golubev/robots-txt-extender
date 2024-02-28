package main

import (
	"fmt"
	"log"
	"net/http"
)

// TODO: add log / slog everywhere
// TODO: check response headers and keep them?

type robotsHandler struct {
	cfg    config
	client http.Client
}

func newRobotsHandler(cfg config) robotsHandler {
	client := http.Client{
		Timeout: cfg.timeoutRobotsRequest,
	}

	return robotsHandler{
		cfg:    cfg,
		client: client,
	}
}

func (rh robotsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalRobotsRequest := &http.Request{
		Method: http.MethodGet,
		URL:    rh.cfg.originalRobotsURL,
		Body:   nil,
	}
	originalRobotsRequest = originalRobotsRequest.WithContext(r.Context())

	originalRobotsResponse, err := rh.client.Do(originalRobotsRequest)
	if err != nil {
		// TODO: double check, that closing response body is not needed here.

		w.WriteHeader(http.StatusBadGateway)
		// TODO: improve error message
		// TODO: log error
		_, _ = fmt.Fprintln(w, "failed to get original robots.txt:", err)
		return
	}

	// TODO: Read and extract original robots.txt with lines from file / content provided in cfg.additionalRobotsFile
	_ = originalRobotsResponse

	//TODO implement me
	panic("implement me")
}

func run() error {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	_ = cfg

	http.Handle(fmt.Sprintf("/%s", cfg.newRobotsEndpoint), newRobotsHandler(cfg))

	// TODO: add prometheus / health check endpoint?

	// TODO: log info line here, about listening on port, etc.
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln("error:", err)
	}
}
