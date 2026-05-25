package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	v := newVerifier()
	if err := v.run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		v.dumpRemoteLogs(os.Stderr)
		os.Exit(1)
	}
}

type verifier struct {
	deployUrl  string
	token      string
	deployHost string
	deployUser string
	keyFile    string
	http       *http.Client
}

func newVerifier() *verifier {
	keyFile := os.Getenv("DEPLOY_KEYFILE")
	if keyFile == "" {
		keyFile = "/tmp/_deploy_key"
	}
	return &verifier{
		deployUrl:  mustEnv("DEPLOY_URL"),
		token:      mustEnv("SYNCLOUD_TOKEN"),
		deployHost: mustEnv("DEPLOY_HOST"),
		deployUser: mustEnv("DEPLOY_USER"),
		keyFile:    keyFile,
		http:       &http.Client{Timeout: 5 * time.Minute},
	}
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		fmt.Fprintf(os.Stderr, "%s is required\n", name)
		os.Exit(1)
	}
	return v
}

func (v *verifier) run() error {
	if err := v.waitForVersion(); err != nil {
		return err
	}
	if err := v.cacheRefresh(); err != nil {
		return err
	}
	if err := v.assertApps(); err != nil {
		return err
	}
	if err := v.assertFind(); err != nil {
		return err
	}
	return v.assertWebUI()
}

func (v *verifier) waitForVersion() error {
	url := v.deployUrl + "/api/ui/v1/version"
	for i := 0; i < 60; i++ {
		resp, err := v.http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				fmt.Println("version endpoint OK")
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("%s did not return 200 after 120s", url)
}

func (v *verifier) cacheRefresh() error {
	url := v.deployUrl + "/syncloud/v1/cache/refresh"
	body := strings.NewReader(fmt.Sprintf(`{"token":%q}`, v.token))
	resp, err := v.http.Post(url, "application/json", body)
	if err != nil {
		return fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("POST %s returned %d (token or aws creds wrong?): %s",
			url, resp.StatusCode, string(raw))
	}
	fmt.Println("cache refresh OK — token and aws creds validated")
	return nil
}

func (v *verifier) assertApps() error {
	url := v.deployUrl + "/api/ui/v1/apps?channel=stable"
	resp, err := v.http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(raw))
	}
	n, err := countJSONList(resp.Body)
	if err != nil {
		return fmt.Errorf("parse %s: %w", url, err)
	}
	if n == 0 {
		return fmt.Errorf("%s returned empty list — cache did not populate", url)
	}
	fmt.Printf("apps OK (%d apps)\n", n)
	return nil
}

func (v *verifier) assertFind() error {
	url := v.deployUrl + "/v2/snaps/find?architecture=amd64&channel=stable"
	resp, err := v.http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s returned %d", url, resp.StatusCode)
	}
	n, err := countResults(resp.Body)
	if err != nil {
		return fmt.Errorf("parse %s: %w", url, err)
	}
	if n == 0 {
		return fmt.Errorf("%s returned no results", url)
	}
	fmt.Printf("snaps/find OK (%d results)\n", n)
	return nil
}

func (v *verifier) assertWebUI() error {
	url := v.deployUrl + "/"
	resp, err := v.http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GET %s returned %d", url, resp.StatusCode)
	}
	fmt.Println("web UI OK")
	return nil
}

func (v *verifier) dumpRemoteLogs(w io.Writer) {
	key, err := os.ReadFile(v.keyFile)
	if err != nil {
		fmt.Fprintf(w, "cannot read %s: %v\n", v.keyFile, err)
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Fprintf(w, "cannot parse %s: %v\n", v.keyFile, err)
		return
	}
	cfg := &ssh.ClientConfig{
		User:            v.deployUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	client, err := ssh.Dial("tcp", v.deployHost+":22", cfg)
	if err != nil {
		fmt.Fprintf(w, "ssh dial %s: %v\n", v.deployHost, err)
		return
	}
	defer client.Close()
	runRemote(client, w, "sudo -n docker ps -a")
	runRemote(client, w, "sudo -n docker logs syncloud-store 2>&1")
}

func runRemote(client *ssh.Client, w io.Writer, cmd string) {
	sess, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(w, "ssh session: %v\n", err)
		return
	}
	defer sess.Close()
	sess.Stdout = w
	sess.Stderr = w
	fmt.Fprintf(w, "\n--- %s ---\n", cmd)
	_ = sess.Run(cmd)
}
