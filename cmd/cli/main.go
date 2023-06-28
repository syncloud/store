package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"net"
	"net/http"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "cli",
	}

	var cmdStart = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh Syncloud Store Cache",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &http.Client{
				Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", api.InternalApi)
					},
				},
			}
			_, err := client.Post("http://unix/refresh", "", nil)
			return err
		},
	}

	rootCmd.AddCommand(cmdStart)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
