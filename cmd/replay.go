package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dk-a-dev/termf1/v2/internal/config"
	"github.com/dk-a-dev/termf1/v2/internal/livetiming"
	"github.com/dk-a-dev/termf1/v2/internal/ui"
	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay <session.jsonl>",
	Short: "Replay a previously recorded live session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		state := livetiming.NewState()
		// Discard logs so we don't clobber the TUI
		logger := log.New(io.Discard, "", 0)

		// Start the server
		srv := livetiming.NewServer(state, logger)
		srv.StartNotifyLoop(5 * time.Second)

		// Watch and notify mechanism from termf1-server
		go func() {
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
		}()

		provider := livetiming.NewReplayProvider(state, filename, logger)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go provider.Run(ctx)

		// Start HTTP Server on an alternate port to avoid collision with live server.
		port := ":8766"
		httpSrv := &http.Server{
			Addr:    port,
			Handler: srv.Handler(),
		}
		go func() {
			_ = httpSrv.ListenAndServe()
		}()
		
		cfg := config.Load()
		cfg.Version = version
		cfg.LiveServerAddr = "http://localhost" + port

		p := tea.NewProgram(
			ui.NewApp(cfg),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running replay: %v\n", err)
			os.Exit(1)
		}
		cancel()
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)
}
