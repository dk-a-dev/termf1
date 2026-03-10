package livetiming

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ── HTTP handlers ─────────────────────────────────────────────────────────────

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"ok":true}`)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s.state.Snapshot())
}

func (s *Server) handleTiming(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	s.state.mu.RLock()
	b, _ := json.Marshal(s.state.Timing)
	s.state.mu.RUnlock()
	w.Write(b)
}

func (s *Server) handleDrivers(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	s.state.mu.RLock()
	b, _ := json.Marshal(s.state.Drivers)
	s.state.mu.RUnlock()
	w.Write(b)
}

func (s *Server) handleWeather(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	s.state.mu.RLock()
	b, _ := json.Marshal(s.state.Weather)
	s.state.mu.RUnlock()
	w.Write(b)
}

func (s *Server) handleRCM(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	s.state.mu.RLock()
	b, _ := json.Marshal(s.state.RaceControlMessages)
	s.state.mu.RUnlock()
	w.Write(b)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.broker.subscribe()
	n := s.sseClients.Add(1)
	s.logger.Printf("[sse] client connected  (total %d)", n)
	defer func() {
		n := s.sseClients.Add(-1)
		s.logger.Printf("[sse] client disconnected (total %d)", n)
		s.broker.unsubscribe(ch)
	}()

	// Send current snapshot immediately on connect.
	writeSSEEvent(w, "state", s.state.Snapshot())
	flusher.Flush()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			writeSSEEvent(w, "state", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeSSEEvent(w http.ResponseWriter, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
