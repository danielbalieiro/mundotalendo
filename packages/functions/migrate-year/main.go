package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mundotalendo/functions/auth"
	"github.com/mundotalendo/functions/types"
)

var (
	dynamoClient *dynamodb.Client
	tableName    string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	tableName = os.Getenv("SST_Resource_DataTable_name")
}

// migrateYear moves every existing reading to the year-aware PK
// "EVENT#LEITURA#<year>" and sets the "year" attribute.
//
// It scans the whole table for anything whose PK begins with "EVENT#LEITURA",
// normalizing every legacy format into the year-aware one:
//   - "EVENT#LEITURA"            (v1.0.9 production items) -> "EVENT#LEITURA#<year>"
//   - "EVENT#LEITURA#<uuid>"     (legacy seed format)      -> "EVENT#LEITURA#<year>"
//   - "EVENT#LEITURA#<year>"     (already migrated)        -> skipped
//
// This is safe to run multiple times: each item is written with its new key
// before the old item is deleted.
func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Starting migration: adding year to existing readings")

	// Validate API key
	apiKey := request.Headers["x-api-key"]
	if apiKey == "" {
		apiKey = request.Headers["X-API-Key"]
	}
	if !auth.ValidateAPIKey(ctx, dynamoClient, apiKey) {
		return errorResponse(401, "UNAUTHORIZED", "Invalid or missing API key"), nil
	}

	// Target year defaults to 2026 (the only year with legacy data)
	year := "2026"
	if y := strings.TrimSpace(request.QueryStringParameters["year"]); y != "" {
		year = y
	}

	migrated, failed, skipped := runMigration(ctx, year)

	response := map[string]interface{}{
		"success":  true,
		"year":     year,
		"migrated": migrated,
		"failed":   failed,
		"skipped":  skipped,
		"message":  fmt.Sprintf("Migration completed: %d items migrated to year %s", migrated, year),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return errorResponse(500, "MIGRATION_ERROR", "Failed to marshal response"), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(responseBody),
	}, nil
}

func runMigration(ctx context.Context, year string) (migrated, failed, skipped int) {
	prefix := "EVENT#LEITURA"
	newPrefix := "EVENT#LEITURA#" + year

	var lastKey map[string]ddbtypes.AttributeValue

	for {
		result, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
			TableName:        aws.String(tableName),
			FilterExpression: aws.String("begins_with(PK, :prefix)"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":prefix": &ddbtypes.AttributeValueMemberS{Value: prefix},
			},
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			log.Printf("Error scanning DynamoDB: %v", err)
			failed++
			break
		}

		for _, item := range result.Items {
			var reading types.LeituraItem
			if err := attributevalue.UnmarshalMap(item, &reading); err != nil {
				log.Printf("WARN: Failed to unmarshal item, skipping: %v", err)
				skipped++
				continue
			}

			oldPK := reading.PK
			oldSK := reading.SK

			// Skip items already migrated to the target year
			if oldPK == newPrefix && reading.Year == year {
				skipped++
				continue
			}

			// Normalize the legacy seed format:
			// PK "EVENT#LEITURA#<uuid>" + SK "COUNTRY#<iso3>"
			// -> PK "EVENT#LEITURA#<year>" + SK "<uuid>#<iso3>#0"
			if strings.HasPrefix(oldSK, "COUNTRY#") {
				uuid := strings.TrimPrefix(oldPK, prefix+"#")
				iso3 := strings.TrimPrefix(oldSK, "COUNTRY#")
				reading.SK = fmt.Sprintf("%s#%s#0", uuid, iso3)
				if reading.WebhookUUID == "" {
					reading.WebhookUUID = uuid
				}
			}

			reading.PK = newPrefix
			reading.Year = year

			av, err := attributevalue.MarshalMap(reading)
			if err != nil {
				log.Printf("WARN: Failed to marshal item: %v", err)
				failed++
				continue
			}

			// Write the new item first (idempotent), then delete the old one.
			if _, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
				TableName: aws.String(tableName),
				Item:      av,
			}); err != nil {
				log.Printf("ERROR: Failed to put item %s#%s: %v", reading.PK, reading.SK, err)
				failed++
				continue
			}

			if _, err := dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key: map[string]ddbtypes.AttributeValue{
					"PK": &ddbtypes.AttributeValueMemberS{Value: oldPK},
					"SK": &ddbtypes.AttributeValueMemberS{Value: oldSK},
				},
			}); err != nil {
				log.Printf("WARN: Failed to delete old item %s#%s: %v", oldPK, oldSK, err)
			}

			migrated++
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	log.Printf("Migration complete: migrated=%d, failed=%d, skipped=%d", migrated, failed, skipped)
	return migrated, failed, skipped
}

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

func main() {
	lambda.Start(handler)
}
