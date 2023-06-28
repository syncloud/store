package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/machine"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "store",
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Syncloud Store",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Default()
			client := rest.New()
			index := storage.New(client, api.Url, machine.DPKGArch, logger)
			public := api.NewSyncloudStore(args[0], index, client, logger)
			internal := api.NewApi(index)
			err := index.Start()
			if err != nil {
				return err
			}
			err = internal.Start()
			if err != nil {
				return err
			}
			return public.Start()
		},
	}

	rootCmd.AddCommand(cmdStart)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
