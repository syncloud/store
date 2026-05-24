package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/sha3"
)

const (
	endpoint = "http://s3"
	bucket   = "test"
	key      = "test"
	secret   = "testtest"
)

func main() {
	dir, err := os.Getwd()
	must(err)
	if base := filepath.Base(dir); base != "test" {
		dir = filepath.Join(dir, "test")
	}

	svc := s3client()
	must(waitReady(svc))

	channels := []string{"master", "stable", "rc"}
	apps := []string{"testapp1", "testapp2"}

	for _, ch := range channels {
		for _, app := range apps {
			must(putFile(svc, fmt.Sprintf("v2/apps/%s/%s/snap.yaml", ch, app),
				filepath.Join(dir, app, "meta", "snap.yaml"), "application/x-yaml"))
			must(putFile(svc, fmt.Sprintf("v2/apps/%s/%s/icon.png", ch, app),
				filepath.Join(dir, "images", app+".png"), "image/png"))
		}
	}

	matches, err := filepath.Glob(filepath.Join(dir, "testapp*.snap"))
	must(err)
	for _, f := range matches {
		name := filepath.Base(f)
		appVer := strings.TrimSuffix(name, ".snap")
		parts := strings.Split(appVer, "_")
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "skip unrecognised snap %q\n", name)
			continue
		}
		app, ver := parts[0], parts[1]

		body, err := os.ReadFile(f)
		must(err)
		h := sha3.New384()
		h.Write(body)
		sha := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		size := fmt.Sprintf("%d", len(body))

		must(put(svc, "apps/"+name, body, "application/octet-stream"))
		must(put(svc, "apps/"+name+".sha384", []byte(sha), "text/plain"))
		must(put(svc, "apps/"+name+".size", []byte(size), "text/plain"))

		rev := fmt.Sprintf(`{"snap-revision":"%s","snap-id":"%s.%s","snap-size":"%s","snap-sha3-385":"%s"}`,
			ver, app, ver, size, sha)
		must(put(svc, "revisions/"+sha+".revision", []byte(rev), "application/json"))
	}

	for _, app := range apps {
		for _, arch := range []string{"amd64", "arm64", "armhf"} {
			must(put(svc, fmt.Sprintf("releases/stable/%s.%s.version", app, arch),
				[]byte("1"), "text/plain"))
		}
	}

	fmt.Println("seed ok")
}

func s3client() *s3.S3 {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("garage"),
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		S3ForcePathStyle: aws.Bool(true),
	}))
	return s3.New(sess)
}

func waitReady(svc *s3.S3) error {
	var lastErr error
	for i := 0; i < 120; i++ {
		_, lastErr = svc.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(bucket)})
		if lastErr == nil {
			fmt.Printf("bucket %q ready after %ds\n", bucket, i)
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("bucket %q not ready after 120s: last error: %v", bucket, lastErr)
}

func putFile(svc *s3.S3, key, path, contentType string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return put(svc, key, body, contentType)
}

func put(svc *s3.S3, key string, body []byte, contentType string) error {
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
