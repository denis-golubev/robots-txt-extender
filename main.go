package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type robotsHandler struct {
	cfg    config
	client http.Client
	logger *slog.Logger
}

func newRobotsHandler(cfg config, logger *slog.Logger) robotsHandler {
	client := http.Client{
		Timeout: cfg.timeoutRobotsRequest,
	}

	return robotsHandler{
		cfg:    cfg,
		client: client,
		logger: logger,
	}
}

func (rh robotsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.logger.Debug("received request")

	originalRobotsRequest := &http.Request{
		Method: http.MethodGet,
		URL:    rh.cfg.originalRobotsURL,
		Body:   nil,
	}
	originalRobotsRequest = originalRobotsRequest.WithContext(r.Context())

	originalRobotsResponse, err := rh.client.Do(originalRobotsRequest)
	if err != nil {
		msg := "failed to get original robots.txt"
		rh.logger.Error(msg, "error", err)

		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintln(w, msg, ":", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			rh.logger.Error("failed to close original robots.txt response body", "error", err)
			// At this point the (error) response may already have been sent, so no response is provided
			// in case of this error.
		}
	}(originalRobotsResponse.Body)

	// TODO: check response headers and keep them?

	rh.logger.Debug("copying original robots.txt to response...")
	// Note: Explicitly not sending 200 OK header here. This is done implicitly if writing
	//       starts successfully.
	_, err = io.Copy(w, originalRobotsResponse.Body)
	if err != nil {
		msg := "failed to read original robots.txt response"
		rh.logger.Error(msg, "error", err)

		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintln(w, msg, ":", err)
		return
	}
	rh.logger.Debug("copying original robots.txt to response successful")

	// Note: According to RFC 9309 (https://www.rfc-editor.org/rfc/rfc9309.html#section-2.2), each line must end
	//       with an EOL. Thus, we can just append the additional robots.txt file to the response. Additional newlines
	//       (for example to delimit a new group) can then be added by the user directly in the file.
	_, err = fmt.Fprintln(w, rh.cfg.additionalRobotsFile)
	if err != nil {
		// Errors at this point can't be reflected back to the client (as we have sent headers above already during the
		// io.Copy call), so we just log them.
		rh.logger.Error("failed to write additional robots.txt to response (additional lines)", "error", err)
		return
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// TODO: make log level configurable from config

	logger.Debug("config loaded",
		"port", cfg.port,
		"originalRobotsURL", cfg.originalRobotsURL,
		"timeoutRobotsRequest", cfg.timeoutRobotsRequest,
		"additionalRobotsFile", cfg.additionalRobotsFile,
		"newRobotsEndpoint", cfg.newRobotsEndpoint,
	)

	http.Handle(fmt.Sprintf("/%s", cfg.newRobotsEndpoint), newRobotsHandler(cfg, logger.WithGroup("newRobotsHandler")))

	// TODO: add prometheus / health check endpoint?

	logger.Info("listening on port", "port", cfg.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Info("starting server")

	if err := run(logger.WithGroup("run")); err != nil {
		log.Fatalln("error:", err)
	}
}
