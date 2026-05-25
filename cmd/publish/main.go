package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/syncloud/store/rest"
)

func main() {
	var storeUrl string
	root := &cobra.Command{Use: "store-publisher"}
	root.PersistentFlags().StringVarP(&storeUrl, "store-url", "s",
		"https://api.store.syncloud.org", "store url")

	var appDir, snapFile, channel, snapYamlPath, iconPath string
	cmdSnap := &cobra.Command{
		Use:   "snap",
		Short: "Upload a snap, snap.yaml and icon for a single arch",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := rest.NewPublishClient(storeUrl)
			if err != nil {
				return err
			}
			p := NewPublisher(client, appDir, snapFile, channel, snapYamlPath, iconPath, os.Stdout)
			return p.Publish()
		},
	}
	cmdSnap.Flags().StringVarP(&appDir, "app-dir", "d", ".", "app source directory; -f/-y/-i are resolved relative to it")
	cmdSnap.Flags().StringVarP(&snapFile, "file", "f", "", "snap file path (default: <app-dir>/<name>_<version>_<arch>.snap derived from snap.yaml and ./version)")
	cmdSnap.Flags().StringVarP(&channel, "channel", "c", "", "channel (master | stable | rc | ...)")
	cmdSnap.Flags().StringVarP(&snapYamlPath, "snap-yaml", "y", "meta/snap.yaml", "path to snap.yaml")
	cmdSnap.Flags().StringVarP(&iconPath, "icon", "i", "meta/gui/icon.png", "path to icon.png")
	_ = cmdSnap.MarkFlagRequired("channel")
	root.AddCommand(cmdSnap)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
