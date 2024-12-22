package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "portsicle",
	Short: "Expose local servers to public network.",
	Long:  "Portsicle is a free and open tool for exposing local servers to public network (the internet).",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}