// Package main implements the SQS consumer Lambda for webhook processing.
//
// This Lambda is triggered by SQS messages containing webhook metadata.
// The actual payload is stored in S3 to avoid SQS message size limits.
//
// Processing flow:
//  1. Parse SQS message to get UUID
//  2. Fetch full payload from S3
//  3. Delete old user readings from DynamoDB
//  4. Process each desafio (country reading)
//  5. Save new readings to DynamoDB
//
// Error handling:
//   - Permanent errors (invalid message, missing payload): Return nil to prevent retry
//   - Transient errors (S3 timeout, DynamoDB throttle): Return error to trigger retry
//   - Partial failures (some countries fail): Log and continue, return nil
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mundotalendo/functions/types"
)

// Config holds the Lambda configuration.
type Config struct {
	TableName  string
	BucketName string
}

// Consumer handles webhook message processing.
type Consumer struct {
	fetcher   *PayloadFetcher
	store     *LeituraStore
	processor *DesafioProcessor
}

// Global consumer instance (initialized in init or lazily on first request)
var consumer *Consumer

// testMode indicates whether we're running in test mode (skips init)
var testMode = false

func init() {
	// Skip initialization in test mode
	if testMode {
		return
	}

	// Check if environment variables are set (allows lazy initialization)
	tableName := os.Getenv("SST_Resource_DataTable_name")
	bucketName := os.Getenv("SST_Resource_PayloadBucket_name")

	if tableName == "" || bucketName == "" {
		// In Lambda, these must be set. Log and exit.
		// In tests, testMode should be true.
		log.Printf("Environment variables not set (test mode or missing config)")
		return
	}

	initConsumer()
}

// initConsumer initializes the consumer with AWS clients.
// Called from init() in production, can be called manually in tests.
func initConsumer() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	tableName := os.Getenv("SST_Resource_DataTable_name")
	bucketName := os.Getenv("SST_Resource_PayloadBucket_name")

	if tableName == "" || bucketName == "" {
		log.Fatalf("missing required environment variables: DataTable=%s, PayloadBucket=%s",
			tableName, bucketName)
	}

	// Initialize clients
	s3Client := s3.NewFromConfig(cfg)
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Initialize consumer components
	fetcher := NewPayloadFetcher(s3Client, bucketName)
	store := NewLeituraStore(dynamoClient, tableName)
	processor := NewDesafioProcessor(store)

	consumer = &Consumer{
		fetcher:   fetcher,
		store:     store,
		processor: processor,
	}

	log.Printf("Consumer initialized: table=%s, bucket=%s", tableName, bucketName)
}

// handler processes SQS events containing webhook messages.
// Each SQS record contains a types.SQSMessage with the webhook UUID.
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("Received %d SQS message(s)", len(sqsEvent.Records))

	for _, record := range sqsEvent.Records {
		if err := consumer.processRecord(ctx, record); err != nil {
			// Return error to trigger SQS retry (if retryable)
			if IsRetryable(err) {
				log.Printf("Retryable error, will retry: %v", err)
				return err
			}
			// Non-retryable error - log and continue (message goes to DLQ after max retries)
			log.Printf("Non-retryable error, skipping: %v", err)
		}
	}

	return nil
}

// processRecord handles a single SQS message.
func (c *Consumer) processRecord(ctx context.Context, record events.SQSMessage) error {
	log.Printf("Processing message ID: %s", record.MessageId)

	// Parse SQS message
	var msg types.SQSMessage
	if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
		log.Printf("ERROR parsing SQS message: %v", err)
		return WrapError("parse_message", "", "", ErrInvalidMessage)
	}

	log.Printf("Processing webhook UUID=%s, User=%s", msg.UUID, msg.User)

	// Fetch payload from S3
	payload, err := c.fetcher.FetchPayload(ctx, msg.UUID)
	if err != nil {
		log.Printf("ERROR fetching payload: %v", err)
		return WrapError("fetch_payload", msg.UUID, "", err)
	}

	// Delete old user readings
	if _, err := c.store.DeleteOldUserReadings(ctx, msg.User); err != nil {
		// Log warning but continue - this is not a fatal error
		log.Printf("WARN: Failed to delete old readings: %v", err)
	}

	// Process desafios
	meta := ProcessingMeta{
		UUID:      msg.UUID,
		User:      payload.Perfil.Nome,
		AvatarURL: payload.Perfil.Imagem,
		Timestamp: parseTimestamp(msg.Timestamp),
	}

	processed, errCount, results := c.processor.ProcessAll(ctx, payload, meta)

	log.Printf("Processing complete: UUID=%s, Processed=%d, Errors=%d",
		msg.UUID, processed, errCount)

	// Log individual results for monitoring
	for _, r := range results {
		if r.Error != nil && !errors.Is(r.Error, ErrCountryNotFound) {
			log.Printf("ERROR processing country %s: %v", r.Country, r.Error)
		}
	}

	// Determine if we should retry
	// Only retry if ALL items failed with a retryable error
	if processed == 0 && errCount > 0 {
		// Check if any error is retryable
		for _, r := range results {
			if r.Error != nil && IsRetryable(r.Error) {
				return WrapError("process_desafios", msg.UUID, "", ErrDynamoDBWrite)
			}
		}
	}

	return nil
}

// parseTimestamp parses an RFC3339 timestamp, returning current time on error.
func parseTimestamp(ts string) time.Time {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Now()
	}
	return t
}

func main() {
	lambda.Start(handler)
}
