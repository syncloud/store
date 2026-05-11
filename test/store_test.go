package test

import (
	"encoding/json"
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

func TestPrepareStore(t *testing.T) {
	arch, err := snapArch()
	assert.NoError(t, err)

	output, err := Ssh("api.store.test", "apt update")
	assert.NoError(t, err, output)
	output, err = Ssh("api.store.test", "apt install -y apache2")
	assert.NoError(t, err, output)

	output, err = Ssh("api.store.test", "/install.sh /store.tar.gz 1 test")
	assert.NoError(t, err, output)

	output, err = Publish("testapp1", 1)
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release promote -n testapp1 -a %s -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Publish("testapp2", 1)
	assert.NoError(t, err, output)

	output, err = Publish("testapp2", 2)
	assert.NoError(t, err, output)

	output, err = Publish("testapp1", 2)
	assert.NoError(t, err, output)
	output, err = Publish("testapp1", 3)
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
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
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

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c master -t %s --store-url http://api.store.test", arch, StoreDir))
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

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
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

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("device", "snap install testapp1")
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)
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

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
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

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 2 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 2 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
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

	output, err := InstallSnapd("/install-snapd-v2.sh /snapd2.tar.gz")
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

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

	output, err := Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp2 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

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

	for i := 0; i < 5; i++ {
		record("testapp1", "testapp1.1", fmt.Sprintf("dev-app1-%d", i))
	}
	for i := 0; i < 2; i++ {
		record("testapp2", "testapp2.1", fmt.Sprintf("dev-app2-%d", i))
	}

	resp, err := client.R().Get("http://api.store.test/api/ui/v1/apps?channel=stable")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

	var apps []struct {
		Name       string `json:"name"`
		SnapID     string `json:"snapId"`
		Popularity int    `json:"popularity"`
	}
	err = json.Unmarshal(resp.Body(), &apps)
	assert.NoError(t, err, string(resp.Body()))

	pop := map[string]int{}
	var order []string
	for _, a := range apps {
		snap := strings.SplitN(a.SnapID, ".", 2)[0]
		pop[snap] = a.Popularity
		order = append(order, snap)
	}
	assert.Equal(t, 5, pop["testapp1"], "ui apps response: %s", string(resp.Body()))
	assert.Equal(t, 2, pop["testapp2"], "ui apps response: %s", string(resp.Body()))

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

	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release set-version -n testapp1 -a %s -v 1 -c stable -t %s --store-url http://api.store.test", arch, StoreDir))
	assert.NoError(t, err, output)

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

func Publish(name string, version int) (string, error) {
	output, err := Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /%s_%d_amd64.snap -b stable -t %s --store-url http://api.store.test", name, version, StoreDir))
	if err != nil {
		return output, err
	}
	output, err = Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /%s_%d_arm64.snap -b stable -t %s --store-url http://api.store.test", name, version, StoreDir))
	if err != nil {
		return output, err
	}
	return Ssh("apps.syncloud.org", fmt.Sprintf("/syncloud-release publish -f /%s_%d_armhf.snap -b stable -t %s --store-url http://api.store.test", name, version, StoreDir))
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
