package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
)

var (
	totalRobotsTxtRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_robots_txt_requests",
		Help: "The total number of requests for the robots.txt file provided by this service.",
	})
	totalRobotsTxtRequestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_robots_txt_request_errors",
		Help: "The total number of errors when serving the robots.txt file. The reason for this is mostly" +
			" the underlying robots.txt not being served properly or network issues. Not all errors may have" +
			" been reflected back to the client, for example if an error occurred after the response has" +
			" been started.",
	})
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
	rh.logger.DebugContext(r.Context(), "received request")

	totalRobotsTxtRequests.Inc()

	originalRobotsRequest := &http.Request{
		Method: http.MethodGet,
		URL:    rh.cfg.originalRobotsURL,
		Body:   nil,
		Header: make(http.Header),
	}
	originalRobotsRequest = originalRobotsRequest.WithContext(r.Context())

	if rh.cfg.xForwardedProto != nil {
		originalRobotsRequest.Header.Set("X-Forwarded-Proto", *rh.cfg.xForwardedProto)
	}

	originalRobotsResponse, err := rh.client.Do(originalRobotsRequest)
	if err != nil {
		totalRobotsTxtRequestErrors.Inc()

		msg := "failed to get original robots.txt"
		rh.logger.ErrorContext(r.Context(), msg, "error", err)

		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintln(w, msg)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			totalRobotsTxtRequestErrors.Inc()

			rh.logger.ErrorContext(r.Context(), "failed to close original robots.txt response body", "error", err)
			// At this point the (error) response may already have been sent, so no response is provided
			// in case of this error.
		}
	}(originalRobotsResponse.Body)

	if rh.cfg.includeOriginalHeaders {
		rh.logger.DebugContext(r.Context(), "copying original headers", "headers", originalRobotsResponse.Header)

		maps.Copy(w.Header(), originalRobotsResponse.Header)

		// Note: Content-Length is not copied, as we are adding additional content to the response.
		w.Header().Del("Content-Length")
	}

	rh.logger.ErrorContext(r.Context(), "copying original robots.txt to response...")
	// Note: Explicitly not sending 200 OK header here. This is done implicitly if writing
	//       starts successfully.
	_, err = io.Copy(w, originalRobotsResponse.Body)
	if err != nil {
		totalRobotsTxtRequestErrors.Inc()

		msg := "failed to read original robots.txt response"
		rh.logger.ErrorContext(r.Context(), msg, "error", err)

		// Note: This may not have any effect if the response has already been started.
		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintln(w, msg)
		return
	}
	rh.logger.ErrorContext(r.Context(), "copying original robots.txt to response successful")

	// Note: According to RFC 9309 (https://www.rfc-editor.org/rfc/rfc9309.html#section-2.2), each line must end
	//       with an EOL. Thus, we can just append the additional robots.txt file to the response. Additional newlines
	//       (for example to delimit a new group) can then be added by the user directly in the file.
	_, err = fmt.Fprintln(w, rh.cfg.additionalRobotsFile)
	if err != nil {
		totalRobotsTxtRequestErrors.Inc()

		// Errors at this point can't be reflected back to the client (as we have sent headers above already during the
		// io.Copy call), so we just log them.
		rh.logger.ErrorContext(r.Context(), "failed to write additional robots.txt to response (additional lines)", "error", err)
		return
	}
}

func run(level *slog.LevelVar, logger *slog.Logger) error {
	cfg, err := loadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	level.Set(cfg.logLevel)

	logger.Debug("config loaded",
		"port", cfg.port,
		"originalRobotsURL", cfg.originalRobotsURL,
		"timeoutRobotsRequest", cfg.timeoutRobotsRequest,
		"additionalRobotsFile", cfg.additionalRobotsFile,
		"newRobotsEndpoint", cfg.newRobotsEndpoint,
		"includeOriginalHeaders", cfg.includeOriginalHeaders,
		"logLevel", cfg.logLevel,
	)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle(fmt.Sprintf("/%s", cfg.newRobotsEndpoint), newRobotsHandler(cfg, logger.WithGroup("newRobotsHandler")))

	logger.Info("server is listening", "port", cfg.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
}

func main() {
	loggerLevel := new(slog.LevelVar)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: loggerLevel,
	}))

	if err := run(loggerLevel, logger); err != nil {
		logger.Error("failed to run server", "err", err)
		os.Exit(1)
	}
}
