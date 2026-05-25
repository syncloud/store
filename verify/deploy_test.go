//go:build integration

package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

var (
	deployUrl  string
	token      string
	deployHost string
	deployUser string
	keyFile    string
	httpClient *http.Client
)

func TestMain(m *testing.M) {
	deployUrl = mustEnv("DEPLOY_URL")
	token = mustEnv("SYNCLOUD_TOKEN")
	deployHost = mustEnv("DEPLOY_HOST")
	deployUser = mustEnv("DEPLOY_USER")
	keyFile = os.Getenv("DEPLOY_KEYFILE")
	if keyFile == "" {
		keyFile = "/tmp/_deploy_key"
	}
	httpClient = &http.Client{Timeout: 5 * time.Minute}
	os.Exit(m.Run())
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		fmt.Fprintf(os.Stderr, "%s is required\n", name)
		os.Exit(1)
	}
	return v
}

func dumpOnFail(t *testing.T) {
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		key, err := os.ReadFile(keyFile)
		if err != nil {
			t.Logf("read %s: %v", keyFile, err)
			return
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			t.Logf("parse %s: %v", keyFile, err)
			return
		}
		cfg := &ssh.ClientConfig{
			User:            deployUser,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		}
		client, err := ssh.Dial("tcp", deployHost+":22", cfg)
		if err != nil {
			t.Logf("ssh dial %s: %v", deployHost, err)
			return
		}
		defer client.Close()
		for _, cmd := range []string{
			"sudo -n docker ps -a",
			"sudo -n docker logs syncloud-store 2>&1",
		} {
			sess, err := client.NewSession()
			if err != nil {
				t.Logf("ssh session: %v", err)
				continue
			}
			out, _ := sess.CombinedOutput(cmd)
			sess.Close()
			t.Logf("--- %s ---\n%s", cmd, string(out))
		}
	})
}

func TestVersion(t *testing.T) {
	dumpOnFail(t)
	url := deployUrl + "/api/ui/v1/version"
	var code int
	for i := 0; i < 60; i++ {
		resp, err := httpClient.Get(url)
		if err == nil {
			code = resp.StatusCode
			_ = resp.Body.Close()
			if code == http.StatusOK {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("%s never returned 200 (last code %d)", url, code)
}

func TestCacheRefresh(t *testing.T) {
	dumpOnFail(t)
	url := deployUrl + "/syncloud/v1/cache/refresh"
	body := strings.NewReader(fmt.Sprintf(`{"token":%q}`, token))
	resp, err := httpClient.Post(url, "application/json", body)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"POST %s returned %d (token or aws creds wrong?): %s",
		url, resp.StatusCode, string(raw))
}

func TestApps(t *testing.T) {
	dumpOnFail(t)
	url := deployUrl + "/api/ui/v1/apps?channel=stable"
	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var apps []json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apps))
	assert.NotEmpty(t, apps, "%s returned empty list — cache did not populate", url)
}

func TestFind(t *testing.T) {
	dumpOnFail(t)
	url := deployUrl + "/v2/snaps/find?architecture=amd64&channel=stable"
	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		Results []json.RawMessage `json:"results"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotEmpty(t, body.Results, "%s returned no results", url)
}

func TestWebUI(t *testing.T) {
	dumpOnFail(t)
	url := deployUrl + "/"
	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
