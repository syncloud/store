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
	"runtime"
	"strings"
	"time"

	"github.com/syncloud/store/crypto"
	"github.com/syncloud/store/model"
)

const partSize int64 = 16 * 1024 * 1024

type PublishClient interface {
	SnapInit(name, version, arch, channel string, size int64, sha384 string, partSize int64) (*model.PublishInitResponse, error)
	SnapPartUrl(key, uploadId string, partNumber int) (string, error)
	SnapFinalise(req model.PublishFinaliseRequest) error
	SnapYaml(name, channel, snapYaml string) error
	Icon(name, channel, iconPngB64 string) error
}

type Publisher struct {
	client       PublishClient
	appDir       string
	snapFile     string
	channel      string
	snapYamlPath string
	iconPath     string
	out          io.Writer
	http         *http.Client
}

func NewPublisher(client PublishClient, appDir, snapFile, channel, snapYamlPath, iconPath string, out io.Writer) *Publisher {
	return &Publisher{
		client:       client,
		appDir:       appDir,
		snapFile:     snapFile,
		channel:      channel,
		snapYamlPath: snapYamlPath,
		iconPath:     iconPath,
		out:          out,
		http:         &http.Client{Timeout: 2 * time.Hour},
	}
}

func (p *Publisher) Publish() error {
	snapYamlPath := resolveAppPath(p.appDir, p.snapYamlPath)
	iconPath := resolveAppPath(p.appDir, p.iconPath)
	snapYaml, err := os.ReadFile(snapYamlPath)
	if err != nil {
		return fmt.Errorf("read snap.yaml: %w", err)
	}
	snapFile := p.snapFile
	if snapFile == "" {
		snapFile, err = deriveSnapFile(p.appDir, snapYaml)
		if err != nil {
			return err
		}
	} else {
		snapFile = resolveAppPath(p.appDir, snapFile)
	}
	name, version, arch, err := parseSnapName(snapFile)
	if err != nil {
		return err
	}
	if err := validateSnapYamlMatches(snapYaml, name); err != nil {
		return err
	}

	if err := p.client.SnapYaml(name, p.channel, string(snapYaml)); err != nil {
		return fmt.Errorf("snap-yaml: %w", err)
	}
	fmt.Fprintln(p.out, "snap.yaml uploaded")

	if icon, ok, ierr := readIcon(iconPath); ierr != nil {
		return fmt.Errorf("read icon: %w", ierr)
	} else if ok {
		if err := p.client.Icon(name, p.channel, icon); err != nil {
			return fmt.Errorf("icon: %w", err)
		}
		fmt.Fprintln(p.out, "icon uploaded")
	}

	return p.uploadSnapBinary(snapFile, name, version, arch)
}

func (p *Publisher) uploadSnapBinary(snapFile, name, version, arch string) error {
	st, err := os.Stat(snapFile)
	if err != nil {
		return err
	}
	size := st.Size()

	sha384, _, err := crypto.SnapFileSHA3_384(snapFile)
	if err != nil {
		return fmt.Errorf("sha3-384: %w", err)
	}

	fmt.Fprintf(p.out, "init: %s %s %s/%s size=%d\n", name, version, arch, p.channel, size)
	init, err := p.client.SnapInit(name, version, arch, p.channel, size, sha384, partSize)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	fmt.Fprintf(p.out, "uploadId=%s parts=%d\n", init.UploadId, init.PartCount)

	parts, err := p.uploadParts(snapFile, init)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	fmt.Fprintln(p.out, "finalise")
	return p.client.SnapFinalise(model.PublishFinaliseRequest{
		Name: name, Version: version, Arch: arch, Channel: p.channel,
		Key: init.Key, UploadId: init.UploadId, Parts: parts,
		Size: size, Sha384: sha384,
	})
}

func (p *Publisher) uploadParts(snapFile string, init *model.PublishInitResponse) ([]model.PublishPart, error) {
	f, err := os.Open(snapFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parts := make([]model.PublishPart, 0, init.PartCount)
	buf := make([]byte, partSize)
	for i := 0; i < init.PartCount; i++ {
		partNumber := i + 1
		n, rerr := io.ReadFull(f, buf)
		if rerr != nil && rerr != io.EOF && rerr != io.ErrUnexpectedEOF {
			return nil, rerr
		}
		body := buf[:n]
		etag, uerr := p.uploadPart(init.PartUrls[i], body)
		if uerr != nil {
			fresh, ferr := p.client.SnapPartUrl(init.Key, init.UploadId, partNumber)
			if ferr != nil {
				return nil, fmt.Errorf("part %d upload failed (%v) and url refresh failed: %w", partNumber, uerr, ferr)
			}
			etag, uerr = p.uploadPart(fresh, body)
			if uerr != nil {
				return nil, fmt.Errorf("part %d upload failed after refresh: %w", partNumber, uerr)
			}
		}
		parts = append(parts, model.PublishPart{PartNumber: partNumber, ETag: etag})
		fmt.Fprintf(p.out, "  part %d/%d uploaded (%d bytes)\n", partNumber, init.PartCount, n)
	}
	return parts, nil
}

func (p *Publisher) uploadPart(url string, body []byte) (string, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.ContentLength = int64(len(body))
	resp, err := p.http.Do(req)
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

var snapNameRe = regexp.MustCompile(`^(?P<Name>.*)_(?P<Version>.*)_(?P<Arch>.*)\.snap$`)

func resolveAppPath(appDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(appDir, path)
}

func debArch(goArch string) string {
	if goArch == "arm" {
		return "armhf"
	}
	return goArch
}

func runtimeGOARCH() string {
	return runtime.GOARCH
}

func snapNameFromYaml(data []byte) (string, error) {
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "name:") {
			n := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			return strings.Trim(n, `"'`), nil
		}
	}
	return "", fmt.Errorf("snap.yaml missing top-level name field")
}

func deriveSnapFile(appDir string, snapYaml []byte) (string, error) {
	name, err := snapNameFromYaml(snapYaml)
	if err != nil {
		return "", err
	}
	versionRaw, err := os.ReadFile(filepath.Join(appDir, "version"))
	if err != nil {
		return "", fmt.Errorf("read version: %w", err)
	}
	version := strings.TrimSpace(string(versionRaw))
	if version == "" {
		return "", fmt.Errorf("version file is empty")
	}
	return filepath.Join(appDir, fmt.Sprintf("%s_%s_%s.snap", name, version, debArch(runtime.GOARCH))), nil
}

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

func validateSnapYamlMatches(snapYaml []byte, expectedName string) error {
	got, err := snapNameFromYaml(snapYaml)
	if err != nil {
		return err
	}
	if got != expectedName {
		return fmt.Errorf("snap.yaml name=%q does not match snap filename name=%q", got, expectedName)
	}
	return nil
}

func readIcon(path string) (string, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	return base64.StdEncoding.EncodeToString(b), true, nil
}
