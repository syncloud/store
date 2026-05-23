package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
)

const partSize int64 = 16 * 1024 * 1024

func main() {
	var storeUrl string
	root := &cobra.Command{Use: "publish"}
	root.PersistentFlags().StringVarP(&storeUrl, "store-url", "s",
		"https://api.store.syncloud.org", "store url")

	var snapFile, channel, snapYamlPath, iconPath string
	cmdSnap := &cobra.Command{
		Use:   "snap",
		Short: "Upload a snap, snap.yaml and icon for a single arch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublish(storeUrl, snapFile, channel, snapYamlPath, iconPath)
		},
	}
	cmdSnap.Flags().StringVarP(&snapFile, "file", "f", "", "snap file path")
	cmdSnap.Flags().StringVarP(&channel, "channel", "c", "", "channel (master | stable | rc | ...)")
	cmdSnap.Flags().StringVarP(&snapYamlPath, "snap-yaml", "y", "meta/snap.yaml", "path to snap.yaml")
	cmdSnap.Flags().StringVarP(&iconPath, "icon", "i", "meta/gui/icon.png", "path to icon.png")
	_ = cmdSnap.MarkFlagRequired("file")
	_ = cmdSnap.MarkFlagRequired("channel")
	root.AddCommand(cmdSnap)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var snapNameRe = regexp.MustCompile(`^(?P<Name>.*)_(?P<Version>.*)_(?P<Arch>.*)\.snap$`)

func parseSnapName(path string) (name, version, arch string, err error) {
	base := filepath.Base(path)
	m := snapNameRe.FindStringSubmatch(base)
	if m == nil {
		return "", "", "", fmt.Errorf("cannot parse snap name from %q (expected <name>_<version>_<arch>.snap)", base)
	}
	return m[snapNameRe.SubexpIndex("Name")],
		m[snapNameRe.SubexpIndex("Version")],
		m[snapNameRe.SubexpIndex("Arch")], nil
}

func runPublish(storeUrl, snapFile, channel, snapYamlPath, iconPath string) error {
	name, version, arch, err := parseSnapName(snapFile)
	if err != nil {
		return err
	}
	st, err := os.Stat(snapFile)
	if err != nil {
		return err
	}
	size := st.Size()

	sha384, _, err := crypto.SnapFileSHA3_384(snapFile)
	if err != nil {
		return fmt.Errorf("sha3-384: %w", err)
	}

	snapYaml, err := os.ReadFile(snapYamlPath)
	if err != nil {
		return fmt.Errorf("read snap.yaml: %w", err)
	}
	if err := validateSnapYamlMatches(snapYaml, name); err != nil {
		return err
	}

	var iconB64 string
	if iconBytes, ierr := os.ReadFile(iconPath); ierr == nil {
		iconB64 = base64.StdEncoding.EncodeToString(iconBytes)
	} else if !errors.Is(ierr, os.ErrNotExist) {
		return fmt.Errorf("read icon: %w", ierr)
	}

	client, err := rest.NewPublishClient(storeUrl)
	if err != nil {
		return err
	}

	fmt.Printf("init: %s %s %s/%s size=%d\n", name, version, arch, channel, size)
	init, err := client.Init(name, version, arch, channel, size, sha384, partSize)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	fmt.Printf("uploadId=%s parts=%d\n", init.UploadId, init.PartCount)

	parts, err := uploadParts(snapFile, init, client)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	fmt.Println("finalise")
	return client.Finalise(model.PublishFinaliseRequest{
		Name: name, Version: version, Arch: arch, Channel: channel,
		Key: init.Key, UploadId: init.UploadId, Parts: parts,
		Size: size, Sha384: sha384,
		SnapYaml:   string(snapYaml),
		IconPngB64: iconB64,
	})
}

func uploadParts(snapFile string, init *model.PublishInitResponse, client *rest.PublishClient) ([]model.PublishPart, error) {
	f, err := os.Open(snapFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parts := make([]model.PublishPart, 0, init.PartCount)
	buf := make([]byte, partSize)
	httpClient := &http.Client{Timeout: 2 * time.Hour}
	for i := 0; i < init.PartCount; i++ {
		partNumber := i + 1
		n, rerr := io.ReadFull(f, buf)
		if rerr != nil && rerr != io.EOF && rerr != io.ErrUnexpectedEOF {
			return nil, rerr
		}
		body := buf[:n]
		etag, uerr := uploadPart(httpClient, init.PartUrls[i], body)
		if uerr != nil {
			fresh, ferr := client.PartUrl(init.Key, init.UploadId, partNumber)
			if ferr != nil {
				return nil, fmt.Errorf("part %d upload failed (%v) and url refresh failed: %w", partNumber, uerr, ferr)
			}
			etag, uerr = uploadPart(httpClient, fresh, body)
			if uerr != nil {
				return nil, fmt.Errorf("part %d upload failed after refresh: %w", partNumber, uerr)
			}
		}
		parts = append(parts, model.PublishPart{PartNumber: partNumber, ETag: etag})
		fmt.Printf("  part %d/%d uploaded (%d bytes)\n", partNumber, init.PartCount, n)
	}
	return parts, nil
}

func uploadPart(c *http.Client, url string, body []byte) (string, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.ContentLength = int64(len(body))
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("PUT %d: %s", resp.StatusCode, string(raw))
	}
	etag := strings.Trim(resp.Header.Get("ETag"), `"`)
	if etag == "" {
		return "", fmt.Errorf("missing ETag")
	}
	return etag, nil
}

func validateSnapYamlMatches(snapYaml []byte, expectedName string) error {
	for _, line := range strings.Split(string(snapYaml), "\n") {
		if strings.HasPrefix(line, "name:") {
			got := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			got = strings.Trim(got, `"'`)
			if got != expectedName {
				return fmt.Errorf("snap.yaml name=%q does not match snap filename name=%q", got, expectedName)
			}
			return nil
		}
	}
	return fmt.Errorf("snap.yaml missing top-level name field")
}
