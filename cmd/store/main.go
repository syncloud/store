package main

import (
	"io/fs"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
	"github.com/syncloud/store/util"
	"github.com/syncloud/store/web"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "store",
	}

	var metricsAddr string
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
			prometheus.MustRegister(popularity)
			ui := api.NewWeb(webFS, cache, popularity)
			iconProxy := api.NewIconProxy(upstream)
			storeServer := api.NewSyncloudStore(listenAddress, cache, client, signer, config.Token, ui, iconProxy, popularity, logger)
			metricsServer := api.NewMetricsServer(metricsAddr, logger)
			internal := api.NewApi(cache)
			if err := cache.Start(); err != nil {
				return err
			}
			if err := internal.Start(); err != nil {
				return err
			}
			if err := metricsServer.Start(); err != nil {
				return err
			}
			if err := storeServer.Start(); err != nil {
				return err
			}
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			return nil
		},
	}
	cmdStart.Flags().StringVar(&metricsAddr, "metrics-addr", ":9090", "address for prometheus /metrics endpoint")

	rootCmd.AddCommand(cmdStart)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
