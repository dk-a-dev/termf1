package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dk-a-dev/termf1/v2/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/v2/internal/ui/styles"
	"github.com/dk-a-dev/termf1/v2/internal/ui/views/schedule"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Print the F1 season schedule",
	Long:  `Fetches and prints the current Formula 1 season schedule.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := jolpica.NewClient()
		races, err := client.GetSchedule(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching schedule: %v\n", err)
			os.Exit(1)
		}

		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || w < 40 {
			w = 80 // fallback width
		}

		fmt.Println(renderSchedule(races, w))
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

// renderSchedule renders all race cards using the shared scheduler package.
func renderSchedule(races []jolpica.Race, width int) string {
	year := schedule.RaceYear(races)
	title := styles.Title.Render(fmt.Sprintf(" 🏁 %s F1 Season Calendar", year))
	sep := styles.Divider.Render(strings.Repeat("─", width))
	
	// CLI defaults to showing the layout without any cursor selection (-1)
	cards := schedule.RenderScheduleCards(races, time.Now(), width, -1)

	return title + "\n" + sep + "\n" + cards
}


