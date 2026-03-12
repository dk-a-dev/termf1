package cmd

import (
	"context"
	"fmt"
	"os"
	"github.com/dk-a-dev/termf1/internal/api/jolpica"
	"github.com/dk-a-dev/termf1/internal/ui/views/standings"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var standingsCmd = &cobra.Command{
	Use:   "standings",
	Short: "Print the F1 championship standings",
	Long:  `Fetches and prints the current Formula 1 Driver and Constructor championship standings.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := jolpica.NewClient()
		ctx := context.Background()

		drivers, err := client.GetDriverStandings(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching driver standings: %v\n", err)
			os.Exit(1)
		}

		constructors, err := client.GetConstructorStandings(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching constructor standings: %v\n", err)
			os.Exit(1)
		}

		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || w < 40 {
			w = 80 // fallback width
		}

		fmt.Println(standings.RenderStandings(drivers, constructors, w))
	},
}

func init() {
	rootCmd.AddCommand(standingsCmd)
}


