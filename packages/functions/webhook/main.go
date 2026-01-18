// Package main implements the webhook endpoint for receiving reading events.
//
// This Lambda receives webhook payloads from Maratona.app and queues them
// for asynchronous processing. The actual data processing happens in the
// consumer Lambda, triggered by SQS.
//
// Processing flow:
//  1. Validate API key
//  2. Validate basic payload structure
//  3. Generate unique UUID
//  4. Save full payload to S3
//  5. Send message to SQS queue
//  6. Return 202 Accepted
//
// Benefits of async processing:
//   - Fast response time (~100ms vs ~2s)
//   - Automatic retries via SQS
//   - Dead letter queue for failed messages
//   - Better scalability and reliability
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/mundotalendo/functions/auth"
	"github.com/mundotalendo/functions/types"
)

// Constants for payload validation
const (
	MaxPayloadSize = 1024 * 1024 // 1 MB
)

// ValidIdentifiers defines the accepted maratona identifiers.
var ValidIdentifiers = map[string]bool{
	"maratona-lendo-paises": true,
	"mundotalendo-2026":     true,
}

// Config holds the Lambda configuration from environment variables.
type Config struct {
	TableName  string
	BucketName string
	QueueURL   string
}

// Webhook handles incoming webhook requests.
type Webhook struct {
	dynamoClient *dynamodb.Client
	s3Client     *s3.Client
	sqsClient    *sqs.Client
	config       Config
}

// Global webhook instance
var webhook *Webhook

// testMode indicates whether we're running in test mode (skips init)
var testMode = false

func init() {
	// Skip initialization in test mode
	if testMode {
		return
	}

	// Check if environment variables are set
	tableName := os.Getenv("SST_Resource_DataTable_name")
	bucketName := os.Getenv("SST_Resource_PayloadBucket_name")
	queueURL := os.Getenv("SST_Resource_WebhookQueue_url")

	if tableName == "" || bucketName == "" || queueURL == "" {
		log.Printf("Environment variables not set (test mode or missing config)")
		return
	}

	initWebhook()
}

// initWebhook initializes the webhook with AWS clients.
func initWebhook() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	webhookConfig := Config{
		TableName:  os.Getenv("SST_Resource_DataTable_name"),
		BucketName: os.Getenv("SST_Resource_PayloadBucket_name"),
		QueueURL:   os.Getenv("SST_Resource_WebhookQueue_url"),
	}

	if webhookConfig.TableName == "" || webhookConfig.BucketName == "" || webhookConfig.QueueURL == "" {
		log.Fatalf("missing required environment variables: TableName=%s, BucketName=%s, QueueURL=%s",
			webhookConfig.TableName, webhookConfig.BucketName, webhookConfig.QueueURL)
	}

	webhook = &Webhook{
		dynamoClient: dynamodb.NewFromConfig(cfg),
		s3Client:     s3.NewFromConfig(cfg),
		sqsClient:    sqs.NewFromConfig(cfg),
		config:       webhookConfig,
	}

	log.Printf("Webhook initialized: table=%s, bucket=%s, queue=%s",
		webhookConfig.TableName, webhookConfig.BucketName, webhookConfig.QueueURL)
}

// handler processes incoming webhook HTTP requests.
func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Received webhook request: %d bytes", len(request.Body))

	// 1. Validate payload size
	if len(request.Body) > MaxPayloadSize {
		log.Printf("Payload too large: %d bytes (max: %d)", len(request.Body), MaxPayloadSize)
		return errorResponse(400, "PAYLOAD_TOO_LARGE", "Payload exceeds 1 MB limit"), nil
	}

	// 2. Validate API key
	apiKey := getAPIKey(request.Headers)
	if !auth.ValidateAPIKey(ctx, webhook.dynamoClient, apiKey) {
		log.Printf("Unauthorized: invalid API key")
		return errorResponse(401, "UNAUTHORIZED", "Invalid or missing API key"), nil
	}

	// 3. Parse and validate payload
	var payload types.WebhookPayload
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		log.Printf("Error parsing payload: %v", err)
		return errorResponse(400, "INVALID_JSON", "Failed to parse JSON payload"), nil
	}

	// 4. Validate identificador
	if !ValidIdentifiers[payload.Maratona.Identificador] {
		log.Printf("Ignoring event with identificador: %s", payload.Maratona.Identificador)
		return successResponse("Event ignored - invalid identificador"), nil
	}

	// 5. Validate required fields
	if payload.Perfil.Nome == "" {
		log.Printf("Validation error: missing perfil.nome")
		return errorResponse(400, "VALIDATION_ERROR", "Missing required field: perfil.nome"), nil
	}

	if len(payload.Desafios) == 0 {
		log.Printf("Validation error: no desafios provided")
		return errorResponse(400, "VALIDATION_ERROR", "No desafios provided"), nil
	}

	// 6. Generate UUID and timestamp
	webhookUUID := uuid.New().String()
	timestamp := time.Now().Format(time.RFC3339)

	log.Printf("Processing webhook UUID=%s for user=%s", webhookUUID, payload.Perfil.Nome)

	// 7. Save payload to S3
	if err := webhook.savePayloadToS3(ctx, webhookUUID, request.Body); err != nil {
		log.Printf("Error saving to S3: %v", err)
		return errorResponse(500, "STORAGE_ERROR", "Failed to store payload"), nil
	}

	// 8. Send message to SQS
	if err := webhook.sendToSQS(ctx, webhookUUID, payload.Perfil.Nome, timestamp); err != nil {
		log.Printf("Error sending to SQS: %v", err)
		// Cleanup S3 on failure
		webhook.deletePayloadFromS3(ctx, webhookUUID)
		return errorResponse(500, "QUEUE_ERROR", "Failed to queue message"), nil
	}

	log.Printf("Webhook queued successfully: UUID=%s, User=%s", webhookUUID, payload.Perfil.Nome)

	// 9. Return 202 Accepted
	return acceptedResponse(webhookUUID), nil
}

// savePayloadToS3 stores the webhook payload in S3.
func (w *Webhook) savePayloadToS3(ctx context.Context, webhookUUID, body string) error {
	key := fmt.Sprintf("payloads/%s.json", webhookUUID)

	_, err := w.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(w.config.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader([]byte(body)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("S3 PutObject failed: %w", err)
	}

	log.Printf("Saved payload to S3: %s/%s", w.config.BucketName, key)
	return nil
}

// deletePayloadFromS3 removes a payload from S3 (used for cleanup on error).
func (w *Webhook) deletePayloadFromS3(ctx context.Context, webhookUUID string) {
	key := fmt.Sprintf("payloads/%s.json", webhookUUID)

	_, err := w.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(w.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("WARN: Failed to cleanup S3 object %s: %v", key, err)
	}
}

// sendToSQS sends a message to the webhook processing queue.
func (w *Webhook) sendToSQS(ctx context.Context, webhookUUID, user, timestamp string) error {
	msg := types.SQSMessage{
		UUID:      webhookUUID,
		User:      user,
		Timestamp: timestamp,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal SQS message failed: %w", err)
	}

	_, err = w.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(w.config.QueueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("SQS SendMessage failed: %w", err)
	}

	log.Printf("Sent message to SQS: UUID=%s", webhookUUID)
	return nil
}

// getAPIKey extracts the API key from request headers.
func getAPIKey(headers map[string]string) string {
	if key := headers["x-api-key"]; key != "" {
		return key
	}
	return headers["X-API-Key"]
}

// Response helpers

func errorResponse(statusCode int, code, message string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]string{
		"error":   code,
		"message": message,
	})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}

func successResponse(message string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"message": message,
	})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}

func acceptedResponse(uuid string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"uuid":    uuid,
		"status":  "QUEUED",
		"message": "Webhook queued for processing",
	})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 202,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}

func main() {
	lambda.Start(handler)
}
