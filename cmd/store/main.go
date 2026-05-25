package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/syncloud/store/api"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/release"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
	"github.com/syncloud/store/util"
	"github.com/syncloud/store/web"
)

func main() {
	var metricsAddr string
	cmdStart := &cobra.Command{
		Use:   "start",
		Short: "Start Syncloud Store",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(args[0], args[1], metricsAddr)
		},
	}
	cmdStart.Flags().StringVar(&metricsAddr, "metrics-addr", ":9090", "address for prometheus /metrics endpoint")

	rootCmd := &cobra.Command{Use: "store"}
	rootCmd.AddCommand(cmdStart)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func start(listenAddress, configPath, metricsAddr string) error {
	logger := log.Default()
	config, err := util.LoadConfig(configPath)
	if err != nil {
		return err
	}
	if config.BaseUrl == "" {
		return fmt.Errorf("base_url is required in %s", configPath)
	}
	if config.Bucket == "" {
		return fmt.Errorf("bucket is required in %s", configPath)
	}
	upstream, err := url.Parse(config.BaseUrl)
	if err != nil {
		return err
	}
	webFS, err := fs.Sub(web.FS, "dist")
	if err != nil {
		return err
	}

	client := rest.New()
	mp, err := release.NewMultipart(config.Bucket, release.AwsConfig{
		AccessKeyId:     config.AwsAccessKeyId,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
		Region:          config.AwsRegion,
	})
	if err != nil {
		return err
	}
	cache := storage.New(client, mp, config.BaseUrl, logger)
	signer := crypto.NewSigner(logger)
	popularity := storage.NewPopularity()
	snapdMetrics := api.NewSnapdMetrics()
	ui := api.NewWeb(webFS, cache, popularity)
	iconProxy := api.NewIconProxy(upstream)
	snapBinary := api.NewSnapBinaryPublisher(mp, cache, config.Token, logger)
	snapYaml := api.NewSnapYamlPublisher(mp, config.Token, logger)
	icon := api.NewIconPublisher(mp, config.Token, logger)
	storeServer := api.NewSyncloudStore(listenAddress, config.BaseUrl, cache, client, signer, config.Token, ui, iconProxy, popularity, snapdMetrics, snapBinary, snapYaml, icon, logger)
	metricsServer := api.NewMetricsServer(metricsAddr, logger, snapdMetrics)
	internal := api.NewApi(cache)

	err = cache.Start()
	if err != nil {
		return err
	}
	err = internal.Start()
	if err != nil {
		return err
	}

	storeErrs := storeServer.Start()
	metricsErrs := metricsServer.Start()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-storeErrs:
		return err
	case err = <-metricsErrs:
		return err
	case <-sig:
		return nil
	}
}
