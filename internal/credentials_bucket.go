package reporter

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type CredentialsBucket struct {
	s3Svc      *s3.S3
	bucketName string
}

func NewCredentialsBucket(awsSess *session.Session) *CredentialsBucket {
	return &CredentialsBucket{
		s3Svc:      s3.New(awsSess),
		bucketName: "listing-reporter",
	}
}

func (r *CredentialsBucket) Get(name string) ([]byte, error) {
	res, err := r.s3Svc.GetObject(&s3.GetObjectInput{
		Bucket: &r.bucketName,
		Key:    &name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", name, err)
	}

	file, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", name, err)
	}

	return file, nil
}
