package main

import (
	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
	"github.com/syncloud/store/util"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "store",
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Syncloud Store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Default()
			config, err := util.LoadConfig(args[1])
			if err != nil {
				return err
			}
			client := rest.New()
			index := storage.New(client, api.Url, logger)
			signer := crypto.NewSigner(logger)
			public := api.NewSyncloudStore(args[0], index, client, signer, config.Token, logger)
			internal := api.NewApi(index)
			err = index.Start()
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
		panic(err)
	}
}
