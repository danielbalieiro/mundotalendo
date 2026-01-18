package main

import (
	"errors"
	"fmt"
)

// Sentinel errors for consumer processing.
// These errors indicate specific failure modes that help determine retry behavior.
var (
	// ErrInvalidMessage indicates the SQS message could not be parsed.
	// This is a permanent failure - do not retry.
	ErrInvalidMessage = errors.New("invalid SQS message format")

	// ErrPayloadNotFound indicates the S3 payload does not exist.
	// This is a permanent failure - do not retry.
	ErrPayloadNotFound = errors.New("payload not found in S3")

	// ErrS3Fetch indicates a transient S3 error occurred.
	// This should trigger a retry.
	ErrS3Fetch = errors.New("failed to fetch payload from S3")

	// ErrInvalidPayload indicates the payload JSON is malformed.
	// This is a permanent failure - do not retry.
	ErrInvalidPayload = errors.New("invalid payload JSON")

	// ErrCountryNotFound indicates a country name could not be mapped to ISO3.
	// This is logged but does not fail the entire message.
	ErrCountryNotFound = errors.New("country not found in mapping")

	// ErrDynamoDBWrite indicates a DynamoDB write operation failed.
	// This should trigger a retry.
	ErrDynamoDBWrite = errors.New("failed to write to DynamoDB")
)

// ProcessingError wraps an error with additional context about the processing failure.
type ProcessingError struct {
	Op      string // Operation that failed (e.g., "fetch_payload", "save_leitura")
	UUID    string // Webhook UUID for tracking
	Country string // Country being processed (if applicable)
	Err     error  // Underlying error
}

// Error implements the error interface.
func (e *ProcessingError) Error() string {
	if e.Country != "" {
		return fmt.Sprintf("%s [uuid=%s, country=%s]: %v", e.Op, e.UUID, e.Country, e.Err)
	}
	return fmt.Sprintf("%s [uuid=%s]: %v", e.Op, e.UUID, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *ProcessingError) Unwrap() error {
	return e.Err
}

// IsRetryable determines if an error should trigger an SQS retry.
// Permanent failures (invalid message, payload not found) should not retry.
// Transient failures (S3 timeouts, DynamoDB throttling) should retry.
func IsRetryable(err error) bool {
	// Permanent failures - do not retry
	if errors.Is(err, ErrInvalidMessage) ||
		errors.Is(err, ErrPayloadNotFound) ||
		errors.Is(err, ErrInvalidPayload) {
		return false
	}

	// Transient failures - retry
	if errors.Is(err, ErrS3Fetch) ||
		errors.Is(err, ErrDynamoDBWrite) {
		return true
	}

	// Default: retry unknown errors
	return true
}

// WrapError creates a ProcessingError with context.
func WrapError(op, uuid, country string, err error) *ProcessingError {
	return &ProcessingError{
		Op:      op,
		UUID:    uuid,
		Country: country,
		Err:     err,
	}
}
