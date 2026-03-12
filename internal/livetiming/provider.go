package livetiming

import "context"

// StreamProvider defines the interface for any backend that populates live f1 data.
// It is expected to block until ctx is cancelled or a fatal error occurs.
type StreamProvider interface {
	Run(ctx context.Context) error
}
