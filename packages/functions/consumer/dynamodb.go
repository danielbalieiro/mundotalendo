package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mundotalendo/functions/types"
)

// DynamoDBClient defines the interface for DynamoDB operations.
// This interface enables mocking in unit tests.
type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// LeituraStore handles DynamoDB operations for reading data.
type LeituraStore struct {
	client    DynamoDBClient
	tableName string
}

// NewLeituraStore creates a new LeituraStore with the given DynamoDB client and table name.
func NewLeituraStore(client DynamoDBClient, tableName string) *LeituraStore {
	return &LeituraStore{
		client:    client,
		tableName: tableName,
	}
}

// SaveLeitura persists a LeituraItem to DynamoDB.
//
// Returns:
//   - error: ErrDynamoDBWrite if the write fails
func (s *LeituraStore) SaveLeitura(ctx context.Context, item types.LeituraItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDynamoDBWrite, err)
	}

	log.Printf("Saved leitura: ISO3=%s, User=%s, Progress=%d%%",
		item.ISO3, item.User, item.Progresso)

	return nil
}

// DeleteOldUserReadings removes all existing readings for a user.
// This ensures each webhook replaces the user's previous data completely.
// It uses the GSI UserIndex to find items efficiently.
//
// Note: This function only deletes EVENT#LEITURA items, not WEBHOOK#PAYLOAD.
//
// Returns:
//   - int: Number of items deleted
//   - error: ErrDynamoDBWrite if deletion fails
func (s *LeituraStore) DeleteOldUserReadings(ctx context.Context, user string) (int, error) {
	log.Printf("Querying old readings for user: %s", user)

	// Query using GSI UserIndex
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("UserIndex"),
		KeyConditionExpression: aws.String("#user = :user"),
		ExpressionAttributeNames: map[string]string{
			"#user": "user",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user": &ddbtypes.AttributeValueMemberS{Value: user},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	if len(result.Items) == 0 {
		log.Printf("No old readings found for user %s", user)
		return 0, nil
	}

	// Delete each EVENT#LEITURA item
	deletedCount := 0
	for _, item := range result.Items {
		pk, okPK := item["PK"].(*ddbtypes.AttributeValueMemberS)
		sk, okSK := item["SK"].(*ddbtypes.AttributeValueMemberS)

		if !okPK || !okSK {
			log.Printf("WARN: Invalid item structure, skipping")
			continue
		}

		// Only delete EVENT#LEITURA items (protect WEBHOOK#PAYLOAD from deletion)
		if !strings.HasPrefix(pk.Value, "EVENT#LEITURA") {
			continue
		}

		_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(s.tableName),
			Key: map[string]ddbtypes.AttributeValue{
				"PK": &ddbtypes.AttributeValueMemberS{Value: pk.Value},
				"SK": &ddbtypes.AttributeValueMemberS{Value: sk.Value},
			},
		})
		if err != nil {
			log.Printf("WARN: Failed to delete %s#%s: %v", pk.Value, sk.Value, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Deleted %d old readings for user %s", deletedCount, user)
	return deletedCount, nil
}
