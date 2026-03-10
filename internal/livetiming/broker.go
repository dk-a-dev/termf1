package livetiming

import "sync"

// sseBroker fans out byte slices to all registered SSE subscriber channels.
type sseBroker struct {
	mu          sync.Mutex
	subscribers map[chan []byte]struct{}
}

func newSSEBroker() *sseBroker {
	return &sseBroker{subscribers: make(map[chan []byte]struct{})}
}

func (b *sseBroker) subscribe() chan []byte {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *sseBroker) unsubscribe(ch chan []byte) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	b.mu.Unlock()
}

// publish broadcasts data to every subscriber, dropping slow consumers.
func (b *sseBroker) publish(data []byte) {
	b.mu.Lock()
	subs := make([]chan []byte, 0, len(b.subscribers))
	for ch := range b.subscribers {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- data:
		default: // slow consumer — drop frame
		}
	}
}
