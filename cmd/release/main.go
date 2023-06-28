package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/release"
	"os"
	"strconv"
)

type SnapRevision struct {
	Revision string `json:"snap-revision"`
	Id       string `json:"snap-id"`
	Size     string `json:"snap-size"`
	Sha384   string `json:"snap-sha3-385"`
}

func main() {

	var rootCmd = &cobra.Command{
		Use: "syncloud-release",
	}

	var target string
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "s3", "target: s3 or local dir")

	var file string
	var branch string
	var storage release.Storage
	var cmdPublish = &cobra.Command{
		Use:   "publish",
		Short: "Publish an app to Syncloud Store",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			sha384, size, err := crypto.SnapFileSHA3_384(file)
			if err != nil {
				return err
			}
			info, err := release.Parse(file, branch)
			if err != nil {
				return err
			}
			storage = NewStorage(target)
			err = storage.UploadFile(file, info.StoreSnapPath)
			if err != nil {
				return err
			}
			err = storage.UploadContent(sha384, info.StoreSha384Path)
			if err != nil {
				return err
			}
			sizeString := strconv.FormatUint(size, 10)
			err = storage.UploadContent(sizeString, info.StoreSizePath)
			if err != nil {
				return err
			}
			err = storage.UploadContent(info.Version, info.StoreVersionPath)
			if err != nil {
				return err
			}
			snapRevision := &SnapRevision{
				Id:       ConstructSnapId(info.Name, info.Version),
				Size:     sizeString,
				Revision: info.Version,
				Sha384:   sha384,
			}
			snapRevisionJson, err := json.Marshal(snapRevision)
			if err != nil {
				return err
			}
			err = storage.UploadContent(string(snapRevisionJson), fmt.Sprintf("revisions/%s.revision", sha384))
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmdPublish.Flags().StringVarP(&file, "file", "f", "", "snap file path")
	err := cmdPublish.MarkFlagRequired("file")
	if err != nil {
		return
	}
	cmdPublish.Flags().StringVarP(&branch, "branch", "b", "", "branch")
	err = cmdPublish.MarkFlagRequired("branch")
	if err != nil {
		return
	}
	rootCmd.AddCommand(cmdPublish)

	var app string
	var arch string
	var cmdPromote = &cobra.Command{
		Use:   "promote",
		Short: "Promote an app to stable channel",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			storage = NewStorage(target)
			version, err := storage.DownloadContent(fmt.Sprintf("releases/rc/%s.%s.version", app, arch))
			if err != nil {
				return err
			}
			return storage.UploadContent(version, fmt.Sprintf("releases/stable/%s.%s.version", app, arch))
		},
	}
	cmdPromote.Flags().StringVarP(&app, "name", "n", "", "app name to promote")
	err = cmdPromote.MarkFlagRequired("name")
	if err != nil {
		return
	}
	cmdPromote.Flags().StringVarP(&arch, "arch", "a", "", "arch to promote")
	err = cmdPromote.MarkFlagRequired("arch")
	if err != nil {
		return
	}
	rootCmd.AddCommand(cmdPromote)

	var channel string
	var version string
	var cmdSetVersion = &cobra.Command{
		Use:   "set-version",
		Short: "Set app version on a channel",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			storage = NewStorage(target)
			return storage.UploadContent(version, fmt.Sprintf("releases/%s/%s.%s.version", channel, app, arch))
		},
	}
	cmdSetVersion.Flags().StringVarP(&app, "name", "n", "", "app")
	err = cmdSetVersion.MarkFlagRequired("name")
	if err != nil {
		return
	}
	cmdSetVersion.Flags().StringVarP(&arch, "arch", "a", "", "arch")
	err = cmdSetVersion.MarkFlagRequired("arch")
	if err != nil {
		return
	}
	cmdSetVersion.Flags().StringVarP(&version, "version", "v", "", "version")
	err = cmdSetVersion.MarkFlagRequired("version")
	if err != nil {
		return
	}
	cmdSetVersion.Flags().StringVarP(&channel, "channel", "c", "", "channel")
	err = cmdSetVersion.MarkFlagRequired("channel")
	if err != nil {
		return
	}
	rootCmd.AddCommand(cmdSetVersion)

	err = rootCmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func NewStorage(target string) release.Storage {
	if target == "s3" {
		return release.NewS3("apps.syncloud.org")
	} else {
		return release.NewFileSystem(target)
	}
}

func ConstructSnapId(name string, version string) string {
	return fmt.Sprintf("%s.%s", name, version)
}
