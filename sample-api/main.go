package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code", "service"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "service"},
	)

	httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"method", "path", "service"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsInFlight)
}

// Response types
type HealthResponse struct {
	Status  string    `json:"status"`
	Time    time.Time `json:"time"`
	Version string    `json:"version"`
}

type K8sClusterResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	NodeCount int               `json:"node_count"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type AuthCheckResponse struct {
	Allowed   bool              `json:"allowed"`
	Resource  string            `json:"resource"`
	Action    string            `json:"action"`
	Subject   string            `json:"subject"`
	Metadata  map[string]string `json:"metadata"`
	CheckedAt time.Time         `json:"checked_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Service configuration
type Service struct {
	logger *logrus.Logger
	port   string
}

func NewService() *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Service{
		logger: logger,
		port:   port,
	}
}

// Prometheus middleware
func (s *Service) prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get route pattern
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		if path == "" {
			path = r.URL.Path
		}

		// Track in-flight requests
		httpRequestsInFlight.WithLabelValues(r.Method, path, "sample-api").Inc()
		defer httpRequestsInFlight.WithLabelValues(r.Method, path, "sample-api").Dec()

		// Wrap response writer to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Add some artificial latency and errors for demo purposes
		s.simulateRealisticBehavior(path, ww)

		next.ServeHTTP(ww, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, path, statusCode, "sample-api").Inc()
		httpRequestDuration.WithLabelValues(r.Method, path, "sample-api").Observe(duration)

		// Log request
		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       path,
			"status":     statusCode,
			"duration":   duration,
			"user_agent": r.UserAgent(),
		}).Info("HTTP request processed")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Simulate realistic API behavior with some latency and errors
func (s *Service) simulateRealisticBehavior(path string, w *responseWriter) {
	// Add random latency
	latency := time.Duration(rand.Intn(500)) * time.Millisecond
	if rand.Float32() < 0.1 { // 10% chance of higher latency
		latency = time.Duration(rand.Intn(2000)+1000) * time.Millisecond
	}
	time.Sleep(latency)

	// Simulate errors (5% chance)
	if rand.Float32() < 0.05 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Simulate client errors (3% chance)
	if rand.Float32() < 0.03 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// Health endpoints
func (s *Service) getLivez(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "alive",
		Time:    time.Now(),
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) getReadyz(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ready",
		Time:    time.Now(),
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Auth endpoints
func (s *Service) authCheck(w http.ResponseWriter, r *http.Request) {
	response := AuthCheckResponse{
		Allowed:   rand.Float32() > 0.2, // 80% success rate
		Resource:  "k8s-cluster",
		Action:    "read",
		Subject:   "user:example@company.com",
		Metadata:  map[string]string{"tenant": "production", "region": "us-east-1"},
		CheckedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) authCheckForUpdate(w http.ResponseWriter, r *http.Request) {
	response := AuthCheckResponse{
		Allowed:   rand.Float32() > 0.3, // 70% success rate for updates
		Resource:  "k8s-cluster",
		Action:    "update",
		Subject:   "user:example@company.com",
		Metadata:  map[string]string{"tenant": "production", "region": "us-east-1"},
		CheckedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// K8s Cluster endpoints
func (s *Service) createK8sCluster(w http.ResponseWriter, r *http.Request) {
	cluster := K8sClusterResponse{
		ID:        fmt.Sprintf("cluster-%d", rand.Intn(10000)),
		Name:      "production-cluster",
		Status:    "creating",
		NodeCount: rand.Intn(10) + 3,
		Labels:    map[string]string{"env": "production", "region": "us-east-1"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cluster)
}

func (s *Service) updateK8sCluster(w http.ResponseWriter, r *http.Request) {
	cluster := K8sClusterResponse{
		ID:        mux.Vars(r)["id"],
		Name:      "production-cluster",
		Status:    "running",
		NodeCount: rand.Intn(10) + 3,
		Labels:    map[string]string{"env": "production", "region": "us-east-1"},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cluster)
}

func (s *Service) deleteK8sCluster(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Policy relationship endpoints
func (s *Service) createPolicyRelationship(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":         fmt.Sprintf("rel-%d", rand.Intn(10000)),
		"policy_id":  fmt.Sprintf("policy-%d", rand.Intn(1000)),
		"cluster_id": fmt.Sprintf("cluster-%d", rand.Intn(1000)),
		"status":     "active",
		"created_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Service) updatePolicyRelationship(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":         fmt.Sprintf("rel-%d", rand.Intn(10000)),
		"policy_id":  fmt.Sprintf("policy-%d", rand.Intn(1000)),
		"cluster_id": fmt.Sprintf("cluster-%d", rand.Intn(1000)),
		"status":     "updated",
		"updated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) deletePolicyRelationship(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Generate some background traffic
func (s *Service) startBackgroundTraffic() {
	go func() {
		urls := []string{
			"http://localhost:8080/api/inventory/v1/livez",
			"http://localhost:8080/api/inventory/v1/readyz",
			"http://localhost:8080/api/inventory/v1beta1/authz/check",
			"http://localhost:8080/api/inventory/v1beta1/resources/k8s-clusters",
		}

		for {
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)

			url := urls[rand.Intn(len(urls))]
			method := "GET"
			if rand.Float32() < 0.3 {
				method = "POST"
			}

			req, _ := http.NewRequest(method, url, nil)
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}
	}()
}

func (s *Service) setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Apply Prometheus middleware
	r.Use(s.prometheusMiddleware)

	// Health endpoints
	r.HandleFunc("/api/inventory/v1/livez", s.getLivez).Methods("GET")
	r.HandleFunc("/api/inventory/v1/readyz", s.getReadyz).Methods("GET")

	// Auth endpoints
	r.HandleFunc("/api/inventory/v1beta1/authz/check", s.authCheck).Methods("POST")
	r.HandleFunc("/api/inventory/v1beta1/authz/checkforupdate", s.authCheckForUpdate).Methods("POST")

	// K8s Cluster endpoints
	r.HandleFunc("/api/inventory/v1beta1/resources/k8s-clusters", s.createK8sCluster).Methods("POST")
	r.HandleFunc("/api/inventory/v1beta1/resources/k8s-clusters/{id}", s.updateK8sCluster).Methods("PUT")
	r.HandleFunc("/api/inventory/v1beta1/resources/k8s-clusters/{id}", s.deleteK8sCluster).Methods("DELETE")

	// Policy relationship endpoints
	r.HandleFunc("/api/inventory/v1beta1/resource-relationships/k8s-policy_is-propagated-to_k8s-cluster", s.createPolicyRelationship).Methods("POST")
	r.HandleFunc("/api/inventory/v1beta1/resource-relationships/k8s-policy_is-propagated-to_k8s-cluster", s.updatePolicyRelationship).Methods("PUT")
	r.HandleFunc("/api/inventory/v1beta1/resource-relationships/k8s-policy_is-propagated-to_k8s-cluster", s.deletePolicyRelationship).Methods("DELETE")

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Health check for Docker
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return r
}

func (s *Service) Start() error {
	router := s.setupRoutes()

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start background traffic after a delay
	go func() {
		time.Sleep(30 * time.Second)
		s.startBackgroundTraffic()
	}()

	// Start server in goroutine
	go func() {
		s.logger.WithField("port", s.port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Server forced to shutdown")
		return err
	}

	s.logger.Info("Server exited")
	return nil
}

func main() {
	service := NewService()
	if err := service.Start(); err != nil {
		logrus.WithError(err).Fatal("Failed to start service")
	}
}
