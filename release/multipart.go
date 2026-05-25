package release

import (
	"bytes"
	"fmt"
	"strings"
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

type AwsConfig struct {
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
	Region          string
}

func NewMultipart(bucket string, aws_ AwsConfig) (*Multipart, error) {
	if aws_.AccessKeyId == "" {
		return nil, fmt.Errorf("aws_access_key_id is not set in secret.yaml")
	}
	if aws_.SecretAccessKey == "" {
		return nil, fmt.Errorf("aws_secret_access_key is not set in secret.yaml")
	}
	region := aws_.Region
	if region == "" {
		region = "us-west-2"
	}
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(aws_.AccessKeyId, aws_.SecretAccessKey, ""),
		Region:      aws.String(region),
	}
	if aws_.Endpoint != "" {
		cfg.Endpoint = aws.String(aws_.Endpoint)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	return &Multipart{bucket: bucket, region: region, svc: s3.New(sess)}, nil
}

func (m *Multipart) Create(key string) (string, error) {
	out, err := m.svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
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
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func (m *Multipart) ListAppIds(channel string) ([]string, error) {
	prefix := fmt.Sprintf("v2/apps/%s/", channel)
	var ids []string
	var token *string
	for {
		out, err := m.svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(m.bucket),
			Prefix:            aws.String(prefix),
			Delimiter:         aws.String("/"),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, cp := range out.CommonPrefixes {
			p := aws.StringValue(cp.Prefix)
			id := p[len(prefix):]
			id = strings.TrimSuffix(id, "/")
			if id != "" {
				ids = append(ids, id)
			}
		}
		if !aws.BoolValue(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}
	return ids, nil
}

