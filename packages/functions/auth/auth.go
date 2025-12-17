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

	// Scan DynamoDB for all active API keys (filter expression has issues, so we validate in code)
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        &tableName,
		FilterExpression: aws.String("begins_with(PK, :pk) AND #active = :active"),
		ExpressionAttributeNames: map[string]string{
			"#active": "active",
		},
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":pk":     &ddbTypes.AttributeValueMemberS{Value: "APIKEY#"},
			":active": &ddbTypes.AttributeValueMemberBOOL{Value: true},
		},
	})

	if err != nil {
		log.Printf("ERROR validating API key: %v", err)
		return false
	}

	// Iterate through results and match the key in code
	for _, item := range result.Items {
		var apiKeyItem APIKeyItem
		if err := attributevalue.UnmarshalMap(item, &apiKeyItem); err != nil {
			log.Printf("ERROR unmarshaling API key item: %v", err)
			continue
		}

		// Check if this key matches and is active
		if apiKeyItem.Key == apiKey && apiKeyItem.Active {
			log.Printf("API key validated successfully: %s", apiKeyItem.Name)
			return true
		}
	}

	log.Printf("API key validation failed: invalid or inactive key")
	return false
}
