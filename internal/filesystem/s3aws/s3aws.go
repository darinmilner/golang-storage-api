package s3aws

import (
	"fileuploader/internal/filesystem"
	"fileuploader/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3 struct {
	Key      string
	Secret   string
	Region   string
	Endpoint string
	Bucket   string
}

func (s *S3) getCredentials() *credentials.Credentials {
	c := credentials.NewStaticCredentials(s.Key, s.Secret, "")
	return c
}

func (s *S3) Put(fileName, folder string) error {
	return nil
}

func (s *S3) List(prefix string) ([]filesystem.Listing, error) {
	var listing []filesystem.Listing

	c := s.getCredentials()
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &s.Endpoint,
		Region:      &s.Region,
		Credentials: c,
	}))

	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				logger.Errorf(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				logger.Errorf(aerr.Error())
			}
		} else {
			logger.Errorf(err.Error())
		}
		return nil, err
	}

	for _, key := range result.Contents {
		b := float64(*key.Size)
		kb := b / 1024
		mb := kb / 1024
		current := filesystem.Listing{
			Etag:         *key.ETag,
			LastModified: *key.LastModified,
			Key:          *key.Key,
			Size:         mb,
		}

		listing = append(listing, current)
	}

	return listing, nil
}

func (s *S3) Delete(itemsToDelete []string) bool {
	return true
}

func (s *S3) Get(destination string, items ...string) error {
	return nil
}
