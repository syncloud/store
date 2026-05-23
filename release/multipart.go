package release

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	PresignedUrlTTL = 24 * time.Hour
	DefaultPartSize = 16 * 1024 * 1024
)

type Multipart struct {
	bucket string
	region string
	svc    *s3.S3
}

func NewMultipart(bucket string) (*Multipart, error) {
	key, ok := os.LookupEnv(AwsKey)
	if !ok {
		return nil, fmt.Errorf("%s env variable is not set", AwsKey)
	}
	secret, ok := os.LookupEnv(AwsSecret)
	if !ok {
		return nil, fmt.Errorf("%s env variable is not set", AwsSecret)
	}
	region := "us-west-2"
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return &Multipart{bucket: bucket, region: region, svc: s3.New(sess)}, nil
}

func (m *Multipart) Create(key string) (string, error) {
	out, err := m.svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(out.UploadId), nil
}

func (m *Multipart) PresignPart(key, uploadId string, partNumber int) (string, error) {
	req, _ := m.svc.UploadPartRequest(&s3.UploadPartInput{
		Bucket:     aws.String(m.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadId),
		PartNumber: aws.Int64(int64(partNumber)),
	})
	return req.Presign(PresignedUrlTTL)
}

func (m *Multipart) Complete(key, uploadId string, parts []*s3.CompletedPart) error {
	_, err := m.svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(m.bucket),
		Key:             aws.String(key),
		UploadId:        aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: parts},
	})
	return err
}

func (m *Multipart) Abort(key, uploadId string) error {
	_, err := m.svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadId),
	})
	return err
}

func (m *Multipart) HeadSize(key string) (int64, error) {
	out, err := m.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}
	return aws.Int64Value(out.ContentLength), nil
}

func (m *Multipart) Put(key string, body []byte, contentType string) error {
	_, err := m.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (m *Multipart) Get(key string) ([]byte, error) {
	out, err := m.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
