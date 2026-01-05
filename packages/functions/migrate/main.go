package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mundotalendo/functions/types"
	"github.com/mundotalendo/functions/utils"
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

// getWebhookPayload retrieves the original webhook payload by UUID
func getWebhookPayload(ctx context.Context, webhookUUID string) (*types.WebhookPayload, error) {
	// Query for the webhook payload
	pk := fmt.Sprintf("WEBHOOK#PAYLOAD#%s", webhookUUID)

	result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: strPtr("PK = :pk"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk": &ddbtypes.AttributeValueMemberS{Value: pk},
		},
		Limit: intPtr(1), // We only need the first one
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query webhook payload: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("webhook payload not found for UUID: %s", webhookUUID)
	}

	// Unmarshal the webhook item
	var webhookItem types.WebhookItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &webhookItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook item: %w", err)
	}

	// Parse the payload JSON
	var payload types.WebhookPayload
	if err := json.Unmarshal([]byte(webhookItem.Payload), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload JSON: %w", err)
	}

	return &payload, nil
}

// extractCapaURL extracts the book cover URL from the webhook payload for a specific country
func extractCapaURL(payload *types.WebhookPayload, pais string) string {
	// Find the desafio matching the country
	for _, desafio := range payload.Desafios {
		// Clean emojis from description (webhook has "📍Brasil", DB has "Brasil")
		cleanedDesc := utils.CleanEmojis(desafio.Descricao)

		if cleanedDesc == pais || desafio.Descricao == pais {
			// Extract capa from vinculados
			for _, vinculado := range desafio.Vinculados {
				if vinculado.Edicao != nil && vinculado.Edicao.Capa != "" {
					return vinculado.Edicao.Capa
				}
			}
		}
	}
	return ""
}

// updateItemCapaURL updates a LeituraItem with the book cover URL
func updateItemCapaURL(ctx context.Context, pk, sk, capaURL string) error {
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: pk},
			"SK": &ddbtypes.AttributeValueMemberS{Value: sk},
		},
		UpdateExpression: strPtr("SET capaURL = :capa"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":capa": &ddbtypes.AttributeValueMemberS{Value: capaURL},
		},
	})
	return err
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Starting migration: populating capaURL for existing readings")

	// Scan all EVENT#LEITURA items
	var allItems []types.LeituraItem
	var lastEvaluatedKey map[string]ddbtypes.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:        &tableName,
			FilterExpression: strPtr("begins_with(PK, :pk)"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":pk": &ddbtypes.AttributeValueMemberS{Value: "EVENT#LEITURA"},
			},
		}

		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := dynamoClient.Scan(ctx, input)
		if err != nil {
			log.Printf("Error scanning DynamoDB: %v", err)
			return errorResponse(500, "Failed to scan DynamoDB"), nil
		}

		// Unmarshal items
		var items []types.LeituraItem
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
			log.Printf("Error unmarshaling items: %v", err)
			return errorResponse(500, "Failed to unmarshal items"), nil
		}

		allItems = append(allItems, items...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	log.Printf("Found %d total readings", len(allItems))

	// Filter items with empty capaURL
	var itemsToMigrate []types.LeituraItem
	for _, item := range allItems {
		if item.CapaURL == "" {
			itemsToMigrate = append(itemsToMigrate, item)
		}
	}

	log.Printf("Found %d readings without capaURL", len(itemsToMigrate))

	// Process each item
	successCount := 0
	failedCount := 0
	skippedCount := 0

	for i, item := range itemsToMigrate {
		log.Printf("Processing %d/%d: User=%s, Country=%s, UUID=%s",
			i+1, len(itemsToMigrate), item.User, item.Pais, item.WebhookUUID)

		// Skip if no webhookUUID
		if item.WebhookUUID == "" {
			log.Printf("  ⚠️  Skipping: No webhookUUID")
			skippedCount++
			continue
		}

		// Get webhook payload
		payload, err := getWebhookPayload(ctx, item.WebhookUUID)
		if err != nil {
			log.Printf("  ❌ Failed to get webhook payload: %v", err)
			failedCount++
			continue
		}

		// Extract capa URL
		capaURL := extractCapaURL(payload, item.Pais)
		if capaURL == "" {
			log.Printf("  ⚠️  No capa found in webhook payload")
			skippedCount++
			continue
		}

		// Update item
		if err := updateItemCapaURL(ctx, item.PK, item.SK, capaURL); err != nil {
			log.Printf("  ❌ Failed to update item: %v", err)
			failedCount++
			continue
		}

		log.Printf("  ✅ Updated with capa: %s", capaURL)
		successCount++
	}

	// Summary
	log.Printf("\n=== MIGRATION SUMMARY ===")
	log.Printf("Total readings: %d", len(allItems))
	log.Printf("Needed migration: %d", len(itemsToMigrate))
	log.Printf("Successfully updated: %d", successCount)
	log.Printf("Failed: %d", failedCount)
	log.Printf("Skipped (no capa): %d", skippedCount)

	response := map[string]interface{}{
		"success":  true,
		"total":    len(allItems),
		"migrated": successCount,
		"failed":   failedCount,
		"skipped":  skippedCount,
		"message":  fmt.Sprintf("Migration completed: %d items updated", successCount),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return errorResponse(500, "Failed to marshal response"), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func errorResponse(statusCode int, message string) events.APIGatewayV2HTTPResponse {
	body := map[string]string{
		"error":   "MIGRATION_ERROR",
		"message": message,
	}
	bodyJSON, _ := json.Marshal(body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(bodyJSON),
	}
}

func main() {
	lambda.Start(handler)
}
