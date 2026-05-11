package main

import (
	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
	"github.com/syncloud/store/util"
	"github.com/syncloud/store/web"
	"io/fs"
	"net/url"
	"time"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "store",
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Syncloud Store",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			listenAddress := args[0]
			configPath := args[1]

			logger := log.Default()
			config, err := util.LoadConfig(configPath)
			if err != nil {
				return err
			}
			upstream, err := url.Parse(api.Url)
			if err != nil {
				return err
			}
			client := rest.New()
			cache := storage.New(client, api.Url, logger)
			signer := crypto.NewSigner(logger)
			webFS, err := fs.Sub(web.FS, "dist")
			if err != nil {
				return err
			}
			popularity := storage.NewPopularity(7 * 24 * time.Hour)
			ui := api.NewWeb(webFS, cache, popularity)
			iconProxy := api.NewIconProxy(upstream)
			public := api.NewSyncloudStore(listenAddress, cache, client, signer, config.Token, ui, iconProxy, popularity, logger)
			internal := api.NewApi(cache)
			err = cache.Start()
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
