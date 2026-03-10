package livetiming

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// Server exposes the aggregated State over HTTP + SSE.
//
// Routes:
//
//	GET /health          → {"ok":true}
//	GET /state           → full JSON snapshot
//	GET /state/timing    → timing map only
//	GET /state/drivers   → driver info map only
//	GET /state/weather   → weather only
//	GET /state/rcm       → race control messages only
//	GET /events          → SSE stream; sends a "state" event on every update
type Server struct {
	state      *State
	logger     *log.Logger
	broker     *sseBroker
	mux        *http.ServeMux
	sseClients atomic.Int64 // live SSE subscriber count
}

// NewServer creates a Server and registers all routes.
func NewServer(state *State, logger *log.Logger) *Server {
	s := &Server{
		state:  state,
		logger: logger,
		broker: newSSEBroker(),
		mux:    http.NewServeMux(),
	}
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/state", s.handleState)
	s.mux.HandleFunc("/state/timing", s.handleTiming)
	s.mux.HandleFunc("/state/drivers", s.handleDrivers)
	s.mux.HandleFunc("/state/weather", s.handleWeather)
	s.mux.HandleFunc("/state/rcm", s.handleRCM)
	s.mux.HandleFunc("/events", s.handleSSE)
	return s
}

// Handler returns the http.Handler wrapped with request logging middleware.
func (s *Server) Handler() http.Handler { return s.requestLogger(s.mux) }

// Notify broadcasts the current state snapshot to all SSE subscribers.
func (s *Server) Notify() {
	snap := s.state.Snapshot()
	s.broker.publish(snap)
	if n := s.sseClients.Load(); n > 0 {
		s.logger.Printf("[sse] broadcast → %d client(s)  snap=%d B", n, len(snap))
	}
}

// StartNotifyLoop periodically calls Notify so SSE clients receive heartbeat
// snapshots even when the SignalR feed is quiet.
func (s *Server) StartNotifyLoop(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			s.Notify()
		}
	}()
}

// ── Middleware ────────────────────────────────────────────────────────────────

// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// requestLogger wraps a handler to log method, path, status, and latency.
func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.logger.Printf("[http] %-4s %-24s %d  %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}


