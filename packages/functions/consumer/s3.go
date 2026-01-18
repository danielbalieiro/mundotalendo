package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mundotalendo/functions/types"
)

// S3Client defines the interface for S3 operations.
// This interface enables mocking in unit tests.
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// PayloadFetcher handles fetching and parsing webhook payloads from S3.
type PayloadFetcher struct {
	client     S3Client
	bucketName string
}

// NewPayloadFetcher creates a new PayloadFetcher with the given S3 client and bucket.
func NewPayloadFetcher(client S3Client, bucketName string) *PayloadFetcher {
	return &PayloadFetcher{
		client:     client,
		bucketName: bucketName,
	}
}

// FetchPayload retrieves a webhook payload from S3 by UUID.
// The payload is stored at payloads/{uuid}.json.
//
// Returns:
//   - *types.WebhookPayload: The parsed payload if successful
//   - error: ErrPayloadNotFound if the object doesn't exist,
//     ErrS3Fetch for transient errors, ErrInvalidPayload if JSON is malformed
func (f *PayloadFetcher) FetchPayload(ctx context.Context, uuid string) (*types.WebhookPayload, error) {
	key := fmt.Sprintf("payloads/%s.json", uuid)

	log.Printf("Fetching payload from S3: bucket=%s, key=%s", f.bucketName, key)

	result, err := f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if the object doesn't exist
		var notFound *s3types.NoSuchKey
		if errors.As(err, &notFound) {
			log.Printf("ERROR Payload not found in S3: %s", key)
			return nil, fmt.Errorf("%w: %s", ErrPayloadNotFound, key)
		}

		log.Printf("ERROR fetching payload from S3: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrS3Fetch, err)
	}
	defer result.Body.Close()

	// Read the body
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("ERROR reading S3 response body: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrS3Fetch, err)
	}

	// Parse JSON
	var payload types.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("ERROR parsing payload JSON: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	log.Printf("Successfully fetched payload for UUID %s, user=%s, desafios=%d",
		uuid, payload.Perfil.Nome, len(payload.Desafios))

	return &payload, nil
}
