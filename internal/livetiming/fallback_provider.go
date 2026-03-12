package livetiming

import (
	"context"
	"log"
)

// FallbackProvider executes multiple stream providers in order.
// If one provider fails (returns a non-nil error), it attempts the next one.
type FallbackProvider struct {
	providers []StreamProvider
	logger    *log.Logger
}

// NewFallbackProvider creates a new Strategy that executes providers sequentially.
func NewFallbackProvider(logger *log.Logger, providers ...StreamProvider) *FallbackProvider {
	return &FallbackProvider{
		providers: providers,
		logger:    logger,
	}
}

// Run executes the providers. It returns nil if a provider completes successfully or contexts are canceled,
// and it will only return an error if ALL providers fail.
func (fp *FallbackProvider) Run(ctx context.Context) error {
	for i, p := range fp.providers {
		fp.logger.Printf("[fallback] starting provider %T (index %d/%d)", p, i+1, len(fp.providers))
		err := p.Run(ctx)
		if err == nil || ctx.Err() != nil {
			return nil // Stopped cleanly or via context cancellation
		}
		fp.logger.Printf("[fallback] provider %T failed: %v", p, err)
	}
	fp.logger.Println("[fallback] all providers failed")
	return nil
}
