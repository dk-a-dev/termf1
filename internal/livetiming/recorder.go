package livetiming

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RecordFrame is exactly what is written as one line to the .jsonl file.
type RecordFrame struct {
	Timestamp time.Time       `json:"time"`
	Raw       json.RawMessage `json:"raw"`
}

// Recorder handles appending raw JSON frames to a file stream.
type Recorder struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
}

// NewRecorder initialises a new .jsonl file in ~/.termf1/data
// The filename is generated automatically unless specified.
func NewRecorder(filename string) (*Recorder, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".termf1", "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if filename == "" {
		filename = fmt.Sprintf("session-%s.jsonl", time.Now().UTC().Format("20060102_150405"))
	}
	path := filepath.Join(dir, filename)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

// Write records a raw SignalR frame with the current timestamp.
func (r *Recorder) Write(raw []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	frame := RecordFrame{
		Timestamp: time.Now(),
		Raw:       raw,
	}
	return r.enc.Encode(frame)
}

// Close closes the underlying file descriptor.
func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}
