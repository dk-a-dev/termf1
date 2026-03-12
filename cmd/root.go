package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version string
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "termf1",
	Short: "A full-featured Formula 1 terminal UI",
	Long: `___________  _______   _______   ___      ___   _______   ____    
("     _   ")/"     "| /"      \ |"  \    /"  | /"     "| /  " \   
 )__/  \\__/(: ______)|:        | \   \  //   |(: ______)/__|| |   
    \\_ /    \/    |  |_____/   ) /\\  \/.    | \/    |     |: |   
    |.  |    // ___)_  //      / |: \.        | // ___)    _\  |   
    \:  |   (:      "||:  __   \ |.  \    /:  |(:  (      /" \_|\  
     \__|    \_______)|__|  \___)|___|\__/|___| \__/     (_______) 
                                                                   

termf1 is a terminal-based Formula 1 dashboard providing live timing,
championship standings, schedule, race analysis`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(appVersion string) {
	version = appVersion
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
}
