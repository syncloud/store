package test

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/uthng/gossh"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	StoreDir = "/var/www/html"
)

func TestUpgrade(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp1_1_%s.snap -b stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release promote -n testapp1 -a %s -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp2_1_%s.snap -b master -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	//output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release promote -n testapp2 -a %s -t %s", arch, StoreDir))
	//assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp1_2_%s.snap -b stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /testapp1_3_%s.snap -b stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = InstallSnapd("/snapd1.tar.gz")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list testapp1")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp1  1        1    stable    syncloud")

	output, err = UpgradeSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap refresh testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap list testapp1")
	assert.NoError(t, err, output)
	assert.Contains(t, output, "testapp1  2        2    latest/stable  syncloud")

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)

}

func TestUnknown(t *testing.T) {
	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap install unknown")
	assert.Error(t, err)
	assert.Contains(t, output, "not found")
}

func TestInstallWarning(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)
	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)
	assert.NotContains(t, output, "Warning")
	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)
}

func TestMasterChannel(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c master -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

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

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

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

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)
	output, err = Ssh("device", "snap refresh testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)

}

func TestRefreshList(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

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

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 2 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

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

	output, err = Ssh("device", "snap refresh --list")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp2")
	assert.NoError(t, err, output)

}

func TestFind(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "/usr/lib/syncloud-store/bin/cli refresh")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap find testapp1")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap find")
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap remove testapp1")
	assert.NoError(t, err, output)

}

func TestRest_SnapsInfo(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := InstallSnapd("/snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s", arch, StoreDir))
	assert.NoError(t, err, output)

	client := resty.New()
	resp, err := client.R().Get("http://device:8080/v2/snaps/info/testapp1?architecture=arm64&fields=architectures")
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

func InstallSnapd(path string) (string, error) {
	output, err := Ssh("device", fmt.Sprintf("/install-snapd.sh %s", path))
	if err != nil {
		return output, err
	}
	return SshWaitFor("device", "snap list",
		func(output string) bool {
			return strings.Contains(output, "No snaps")
		},
	)
}

func UpgradeSnapd(path string) (string, error) {
	return Ssh("device", fmt.Sprintf("/upgrade-snapd.sh %s", path))
}

func SshWaitFor(host string, command string, predicate func(string) bool) (string, error) {
	retries := 10
	retry := 0
	for retry < retries {
		retry++
		output, err := Ssh(host, command)
		if err != nil {
			fmt.Printf("error: %v", err)
			time.Sleep(1 * time.Second)
			fmt.Printf("retry %d/%d", retry, retries)
			continue
		}
		if predicate(output) {
			return output, nil
		}
	}
	return "", fmt.Errorf("%d: %d (exhausted)", retry, retries)
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
	output, err := client.ExecCommand(command)
	result := string(output)
	fmt.Printf("output: \n%s\n", result)
	return result, err
}
