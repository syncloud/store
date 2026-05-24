package test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/uthng/gossh"
)

const (
	S3Endpoint = "http://s3"
	MinioAccess   = "GK31c4cef60f8f78b1bf12cd71"
	MinioSecret   = "b8a31bf6c5d4e7a9f2b3c1d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8"
	Bucket        = "s3"
)

func TestPrepareStore(t *testing.T) {
	output, err := Ssh("api.store.test", "apt update")
	assert.NoError(t, err, output)
	output, err = Ssh("api.store.test", "apt install -y apache2")
	assert.NoError(t, err, output)

	output, err = Ssh("api.store.test", "/install.sh /store.tar.gz 1 test")
	assert.NoError(t, err, output)
}

func TestUnknown(t *testing.T) {
	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap install unknown")
	assert.Error(t, err)
	assert.Contains(t, output, "not found")
}

func TestInstallWarning(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)
	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)
	assert.NotContains(t, output, "Warning")
	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestMasterChannel(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "master"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap install testapp1 --channel=master")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list testapp1")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp1  1        1    master/stable  syncloud")

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestCommand(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap run testapp1.test")
	assert.NoError(t, err, output)
	assert.NotContains(t, output, "error")

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestRefresh(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "2", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap refresh testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestRefreshList(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, SetVersion("testapp2", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap install testapp2")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list testapp1")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp1  1        1    latest/stable  syncloud")

	output, err = Ssh("device", "snap list testapp2")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp2  1        1    latest/stable  syncloud")

	assert.NoError(t, SetVersion("testapp1", arch, "2", "stable"))
	assert.NoError(t, SetVersion("testapp2", arch, "2", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap refresh --list")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list testapp1")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp1  2        2    latest/stable  syncloud")

	output, err = Ssh("device", "snap list testapp2")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp2  2        2    latest/stable  syncloud")

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp2")
	assert.NoError(t, err, output)
}

func TestFind(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, SetVersion("testapp2", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	output, err = Ssh("device", "snap find testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap find")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestPopularityRanking(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, SetVersion("testapp2", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	client := resty.New()

	record := func(name, snapId, deviceId string) {
		body := fmt.Sprintf(`{"actions":[{"action":"refresh","instance-key":"k","name":"%s","snap-id":"%s","channel":"stable"}]}`, name, snapId)
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Syncloud-Architecture", arch).
			SetHeader("Syncloud-Device-Id", deviceId).
			SetBody(body).
			Post("http://api.store.test/v2/snaps/refresh")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))
	}

	type uiApp struct {
		Name       string `json:"name"`
		SnapID     string `json:"snapId"`
		Popularity int    `json:"popularity"`
	}
	read := func() (map[string]int, []string) {
		resp, err := client.R().Get("http://api.store.test/api/ui/v1/apps?channel=stable")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))
		var apps []uiApp
		assert.NoError(t, json.Unmarshal(resp.Body(), &apps), string(resp.Body()))
		pop := map[string]int{}
		var order []string
		for _, a := range apps {
			snap := strings.SplitN(a.SnapID, ".", 2)[0]
			pop[snap] = a.Popularity
			order = append(order, snap)
		}
		return pop, order
	}

	before, _ := read()

	for i := 0; i < 5; i++ {
		record("testapp1", "testapp1.1", fmt.Sprintf("dev-app1-%d", i))
	}
	for i := 0; i < 2; i++ {
		record("testapp2", "testapp2.1", fmt.Sprintf("dev-app2-%d", i))
	}

	after, order := read()

	assert.Equal(t, 5, after["testapp1"]-before["testapp1"], "delta testapp1 mismatch; before=%v after=%v", before, after)
	assert.Equal(t, 2, after["testapp2"]-before["testapp2"], "delta testapp2 mismatch; before=%v after=%v", before, after)

	idx1, idx2 := -1, -1
	for i, n := range order {
		if n == "testapp1" {
			idx1 = i
		}
		if n == "testapp2" {
			idx2 = i
		}
	}
	assert.True(t, idx1 >= 0 && idx2 >= 0 && idx1 < idx2, "testapp1 should rank before testapp2; got order %v", order)
}

func TestRest_SnapsInfo(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	assert.NoError(t, SetVersion("testapp1", arch, "1", "stable"))
	assert.NoError(t, RefreshCache())

	client := resty.New()
	resp, err := client.R().Get(fmt.Sprintf("http://api.store.test/v2/snaps/info/testapp1?architecture=%s&fields=architectures", arch))
	assert.NoError(t, err, output)
	assert.Equal(t, 200, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), `"snap-id":"testapp1.1"`)
}

func snapArch() (string, error) {
	output, err := exec.Command("dpkg", "--print-architecture").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func InstallSnapd(cmd string) (string, error) {
	output, err := Ssh("device", cmd)
	if err != nil {
		return output, err
	}
	output, err = SshWaitFor("device", "snap list",
		func(output string) bool {
			return strings.Contains(output, "No snaps")
		},
	)
	if err != nil {
		return output, err
	}
	return Ssh("device", "snap wait system seed.loaded")
}

func SshWaitFor(host string, command string, predicate func(string) bool) (string, error) {
	retries := 60
	for retry := 1; retry <= retries; retry++ {
		output, err := Ssh(host, command)
		if err == nil && predicate(output) {
			return output, nil
		}
		fmt.Printf("retry %d/%d: err=%v\n", retry, retries, err)
		time.Sleep(1 * time.Second)
	}
	return "", fmt.Errorf("waited %d retries for %q on %s", retries, command, host)
}

func Ssh(host string, command string) (string, error) {
	config, err := gossh.NewClientConfigWithUserPass("root", "syncloud", host, 22, false)
	if err != nil {
		return "", err
	}

	client, err := gossh.NewClient(config)
	if err != nil {
		return "", err
	}
	fmt.Printf("%s: %s\n", host, command)
	output, err := client.ExecCommand(fmt.Sprintf("SYNCLOUD_TOKEN=test %s", command))
	result := string(output)
	fmt.Printf("output: \n%s\n", result)
	return result, err
}

var s3svc *s3.S3

func s3client() *s3.S3 {
	if s3svc != nil {
		return s3svc
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:         aws.String(S3Endpoint),
		Region:           aws.String("garage"),
		Credentials:      credentials.NewStaticCredentials(MinioAccess, MinioSecret, ""),
		S3ForcePathStyle: aws.Bool(true),
	}))
	s3svc = s3.New(sess)
	return s3svc
}

func SetVersion(app, arch, version, channel string) error {
	_, err := s3client().PutObject(&s3.PutObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(fmt.Sprintf("releases/%s/%s.%s.version", channel, app, arch)),
		Body:   strings.NewReader(version),
	})
	return err
}

func RefreshCache() error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"token":"test"}`).
		Post("http://api.store.test/syncloud/v1/cache/refresh")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("refresh failed: %d %s", resp.StatusCode(), resp.String())
	}
	return nil
}
