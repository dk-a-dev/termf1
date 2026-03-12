package livetiming

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// ReplayProvider implements StreamProvider by reading a .jsonl recording
// and dispatching the frames with the same timing delays as they were recorded.
type ReplayProvider struct {
	filename string
	client   *Client // used for its routing/dispatching logic into state
	logger   *log.Logger
}

// NewReplayProvider creates a replay provider reading from filename.
func NewReplayProvider(state *State, filename string, logger *log.Logger) *ReplayProvider {
	return &ReplayProvider{
		filename: filename,
		client:   NewClient(state, logger),
		logger:   logger,
	}
}

// Run executes the replay, blocking until EOF or context cancellation.
func (r *ReplayProvider) Run(ctx context.Context) error {
	f, err := os.Open(r.filename)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	defer f.Close()

	r.logger.Printf("[replay] playing back %s", r.filename)

	scanner := bufio.NewScanner(f)

	// Max capacity for potentially large SignalR frames
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	var lastTime time.Time

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil
		}

		var frame RecordFrame
		if err := json.Unmarshal(scanner.Bytes(), &frame); err != nil {
			r.logger.Printf("[replay] skipping bad frame: %v", err)
			continue
		}

		if !lastTime.IsZero() && frame.Timestamp.After(lastTime) {
			delay := frame.Timestamp.Sub(lastTime)
			// Cap huge delays (e.g. if the recorder computer slept)
			if delay > 2*time.Minute {
				delay = 2 * time.Second
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil
			}
		}

		lastTime = frame.Timestamp
		r.client.Dispatch(frame.Raw)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("replay read error: %w", err)
	}

	r.logger.Println("[replay] playback finished")
	return nil
}
