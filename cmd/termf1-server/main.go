package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devkeshwani/termf1/internal/livetiming"
)

func main() {
	addr := flag.String("addr", ":8765", "HTTP listen address")
	flag.Parse()

	logger := log.New(os.Stderr, "[termf1-server] ", log.LstdFlags)

	state := livetiming.NewState()
	srv := livetiming.NewServer(state, logger)

	// Start SSE heartbeat every 5 s so connected clients stay alive.
	srv.StartNotifyLoop(5 * time.Second)

	// Wrap the state's apply methods with a notify call.
	// We do this by patching state updates to call srv.Notify() after each
	// write — handled via a goroutine that watches UpdatedAt.
	go watchAndNotify(state, srv)

	// Start the SignalR client.
	client := livetiming.NewClient(state, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.Run(ctx)

	// Start HTTP server.
	httpSrv := &http.Server{
		Addr:         *addr,
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE streams are long-lived
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s", *addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("http: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("shutting down...")
	cancel()
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
	fmt.Println("bye")
}

// watchAndNotify polls state.UpdatedAt and fires srv.Notify() whenever it
// changes. This decouples the state package from the server without adding
// callback indirection inside the state methods.
func watchAndNotify(state *livetiming.State, srv *livetiming.Server) {
	var last time.Time
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for range t.C {
		state.RLock()
		updated := state.UpdatedAt
		state.RUnlock()
		if updated.After(last) {
			last = updated
			srv.Notify()
		}
	}
}
