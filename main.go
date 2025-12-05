package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
	expand "github.com/openvenues/gopostal/expand"
	parser "github.com/openvenues/gopostal/parser"
)

type Request struct {
	Query string   `json:"query"`
	Langs []string `json:"langs"`
}

var (
	promEnabled            bool
	promPort               string
	latencyBuckets         []float64
	expandRequestCtr       prometheus.Counter
	expandLatencyHist      prometheus.Histogram
	parseRequestCtr        prometheus.Counter
	parseLatencyHist       prometheus.Histogram
	expandParseRequestCtr  prometheus.Counter
	expandParseLatencyHist prometheus.Histogram
)

func initPrometheus() {
	latencyBuckets = []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 5.0, 10.0}
	expandRequestCtr = promauto.NewCounter(prometheus.CounterOpts{
		Name: "libpostal_expand_reqs_total",
		Help: "The total number of processed expand requests",
	})
	expandLatencyHist = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "libpostal_expand_durations_historgram_ms",
		Help:    "Latency distributions of expand calls in milliseconds",
		Buckets: latencyBuckets,
	})
	parseRequestCtr = promauto.NewCounter(prometheus.CounterOpts{
		Name: "libpostal_parse_reqs_total",
		Help: "The total number of processed parse requests",
	})
	parseLatencyHist = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "libpostal_parse_durations_historgram_ms",
		Help:    "Latency distributions of parse calls in milliseconds",
		Buckets: latencyBuckets,
	})
	expandParseRequestCtr = promauto.NewCounter(prometheus.CounterOpts{
		Name: "libpostal_expandparse_reqs_total",
		Help: "The total number of processed expandparse requests",
	})
	expandParseLatencyHist = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "libpostal_expandparse_durations_historgram_ms",
		Help:    "Latency distributions of expandparse calls in milliseconds",
		Buckets: latencyBuckets,
	})
}

func startTrace(hist prometheus.Histogram) (prometheus.Histogram, time.Time) {
	return hist, time.Now()
}

func endTrace(hist prometheus.Histogram, startTime time.Time) {
	endTime := time.Now()
	elapsed := float64(endTime.Sub(startTime).Nanoseconds()) / 1000000.0
	if promEnabled {
		hist.Observe(elapsed)
	}
}

func counterInc(counter prometheus.Counter) {
	if promEnabled {
		counter.Inc()
	}
}

func main() {
	host := os.Getenv("LISTEN_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logStructured := os.Getenv("LOG_STRUCTURED")
	if logStructured == "" {
		logStructured = "false"
	}
	jsonLog, err := strconv.ParseBool(logStructured)
	if err != nil {
		jsonLog = false
	}
	if !jsonLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	zLevel, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Msgf("Unknown log level provided: %s", logLevel)
		log.Warn().Msg("Using info level by default")
		zLevel = zerolog.InfoLevel
	}
	log.Info().Msgf("setting log level to '%s'", logLevel)
	zerolog.SetGlobalLevel(zLevel)
	// Checking for flag to enable Prometheus metrics collection
	promPort := os.Getenv("PROMETHEUS_PORT")
	if promPort == "" {
		promPort = "9090"
	}
	promFlag := os.Getenv("PROMETHEUS_ENABLED")
	if promFlag == "" {
		promFlag = "false"
	}
	promEnabled, err = strconv.ParseBool(promFlag)
	if err != nil {
		log.Warn().Msgf("Expected boolean in environment variable 'PROMETHEUS_ENABLED' but got %s", promFlag)
		promEnabled = false
	}
	listenSpec := fmt.Sprintf("%s:%s", host, port)

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	router := mux.NewRouter()
	router.HandleFunc("/health", HealthHandler).Methods("GET")
	router.HandleFunc("/expand", ExpandHandler).Methods("POST")
	router.HandleFunc("/parser", ParserHandler).Methods("POST")
	router.HandleFunc("/expandparser", ExpandParserHandler).Methods("POST")

	var promEndpoint *http.Server

	if promEnabled {
		log.Info().Msg("Prometheus metrics collector and endpoint enabled!")
		initPrometheus()
		promRouter := mux.NewRouter()
		promRouter.Handle("/metrics", promhttp.Handler())
		promListener := fmt.Sprintf("%s:%s", host, promPort)

		log.Info().Msg("Starting Prometheus metrics endpoint")
		promEndpoint = &http.Server{Addr: promListener, Handler: promRouter}

		go func() {
			log.Info().Msgf("metrics endpoint listening on http://%s", promListener)
			if err := promEndpoint.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("Prometheus endpoint error")
			}
		}()
	}

	log.Info().Msg("Starting libpostal-rest service")

	s := &http.Server{Addr: listenSpec, Handler: router}
	go func() {
		var err error
		if certFile != "" && keyFile != "" {
			log.Info().Msgf("listening on https://%s", listenSpec)
			err = s.ListenAndServeTLS(certFile, keyFile)
		} else {
			log.Info().Msgf("listening on http://%s", listenSpec)
			err = s.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Server error")
		}
	}()

	stop := make(chan os.Signal, 1)
	// reacting to Interrupt (CTRL+C) or TERM signals to gracefully terminate the process
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Info().Msg("Shut down signal received")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()

	if err := s.Shutdown(ctx1); err != nil {
		log.Error().Err(err).Msg("Error shutting down main server")
	}
	if promEnabled {
		defer cancel2()
		if err := promEndpoint.Shutdown(ctx2); err != nil {
			log.Error().Err(err).Msg("Error shutting down Prometheus endpoint")
		}
	}
	log.Info().Msg("Server stopped")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling '/expand' request")
	defer endTrace(startTrace(expandLatencyHist))
	w.Header().Set("Content-Type", "application/json")

	var req Request

	q, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(q, &req); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var expansions []string = nil
	if len(req.Langs) > 0 {
		options := expand.GetDefaultExpansionOptions()
		options.Languages = req.Langs
		expansions = expand.ExpandAddressOptions(req.Query, options)
	} else {
		expansions = expand.ExpandAddress(req.Query)
	}
	expansionThing, err := json.Marshal(expansions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal expansions")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	counterInc(expandRequestCtr)
	w.WriteHeader(http.StatusOK)
	w.Write(expansionThing)
}

func ParserHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling '/parser' request")
	defer endTrace(startTrace(parseLatencyHist))
	w.Header().Set("Content-Type", "application/json")

	var req Request

	q, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(q, &req); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	parsed := parser.ParseAddress(req.Query)
	parseThing, err := json.Marshal(parsed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal parsed address")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	counterInc(parseRequestCtr)
	w.WriteHeader(http.StatusOK)
	w.Write(parseThing)
}

func ExpandParserHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling '/expandparser' request")
	defer endTrace(startTrace(expandParseLatencyHist))
	w.Header().Set("Content-Type", "application/json")

	var req Request

	q, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(q, &req); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	inputQuery := req.Query
	expansionsParsed := []map[string]interface{}{{
		"type":   "query",
		"data":   inputQuery,
		"parsed": parser.ParseAddress(inputQuery),
	}}

	var expansions []string = nil
	if len(req.Langs) > 0 {
		options := expand.GetDefaultExpansionOptions()
		options.Languages = req.Langs
		expansions = expand.ExpandAddressOptions(req.Query, options)
	} else {
		expansions = expand.ExpandAddress(req.Query)
	}

	for _, elem := range expansions {
		expansionsParsed = append(expansionsParsed, map[string]interface{}{
			"type":   "expansion",
			"data":   elem,
			"parsed": parser.ParseAddress(elem),
		})
	}

	expansionParserThing, err := json.Marshal(expansionsParsed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal expansion parser result")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	counterInc(expandParseRequestCtr)
	w.WriteHeader(http.StatusOK)
	w.Write(expansionParserThing)
}
