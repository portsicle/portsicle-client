package cmd

import (
	"log"

	"github.com/portsicle/portsicle-client/cmd/client"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "http",
	Short: "Expose local http port",
	Run:   initHTTPClient,
}

func initHTTPClient(cmd *cobra.Command, args []string) {
	httpPort, err := cmd.Flags().GetString("port")
	if err != nil {
		log.Fatalf("Error retrieving port flag: %v", err)
	}

	client.HandleClient(httpPort)
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("port", "p", "8888", "Port on which your local serve is listening.")
}
