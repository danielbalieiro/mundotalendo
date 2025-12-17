package auth

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type APIKeyItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	Name      string `dynamodbav:"name"`
	Key       string `dynamodbav:"key"`
	CreatedAt string `dynamodbav:"createdAt"`
	Active    bool   `dynamodbav:"active"`
}

// ValidateAPIKey checks if the provided API key is valid and active
func ValidateAPIKey(ctx context.Context, client *dynamodb.Client, apiKey string) bool {
	if apiKey == "" {
		log.Printf("API key validation failed: empty key")
		return false
	}

	tableName := os.Getenv("SST_Resource_DataTable_name")
	if tableName == "" {
		log.Printf("ERROR: SST_Resource_DataTable_name is empty")
		return false
	}

	// Scan DynamoDB to find the API key
	// Note: In production, consider adding a GSI on the 'key' attribute for better performance
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        &tableName,
		FilterExpression: aws.String("begins_with(PK, :pk) AND #key = :apikey AND #active = :active"),
		ExpressionAttributeNames: map[string]string{
			"#key":    "key",
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":pk":     &ddbTypes.AttributeValueMemberS{Value: "APIKEY#"},
			":apikey": &ddbTypes.AttributeValueMemberS{Value: apiKey},
			":active": &ddbTypes.AttributeValueMemberBOOL{Value: true},
		},
		Limit: aws.Int32(1), // We only need to know if at least one exists
	})

	if err != nil {
		log.Printf("ERROR validating API key: %v", err)
		return false
	}

	// If we found at least one matching active key, it's valid
	if len(result.Items) > 0 {
		var item APIKeyItem
		if err := attributevalue.UnmarshalMap(result.Items[0], &item); err == nil {
			log.Printf("API key validated successfully: %s", item.Name)
			return true
		}
	}

	log.Printf("API key validation failed: invalid or inactive key")
	return false
}
