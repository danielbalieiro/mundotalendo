package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Fetching user locations from DynamoDB")

	// Validate API key
	apiKey := request.Headers["x-api-key"]
	if apiKey == "" {
		apiKey = request.Headers["X-API-Key"]
	}
	if !auth.ValidateAPIKey(ctx, dynamoClient, apiKey) {
		log.Printf("Unauthorized: invalid API key")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error":"UNAUTHORIZED","message":"Invalid or missing API key"}`,
		}, nil
	}

	// Query DynamoDB for all readings with pagination
	var allItems []map[string]ddbTypes.AttributeValue
	var lastKey map[string]ddbTypes.AttributeValue

	for {
		result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              &tableName,
			KeyConditionExpression: aws.String("PK = :pk"),
			ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
				":pk": &ddbTypes.AttributeValueMemberS{Value: "EVENT#LEITURA"},
			},
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			log.Printf("Error querying DynamoDB: %v", err)
			return errorResponse(500, "Error fetching data"), nil
		}

		allItems = append(allItems, result.Items...)

		// Check if there are more pages
		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	log.Printf("Fetched %d total items from DynamoDB", len(allItems))

	// Find most recent reading per user
	userLatest := make(map[string]types.LeituraItem) // user -> latest item

	for _, item := range allItems {
		var reading types.LeituraItem
		err := attributevalue.UnmarshalMap(item, &reading)
		if err != nil {
			log.Printf("Error unmarshaling item: %v", err)
			continue
		}

		// Skip if user name is empty
		if reading.User == "" {
			continue
		}

		// Skip if progress is 0% (v1.0.3: GPS markers only for active readings)
		if reading.Progresso < 1 {
			continue
		}

		// Keep latest reading per user (v1.0.3: use UpdatedAt timestamp, not SK)
		// SK now is <uuid>#<iso3>#<index> - doesn't reflect temporal order
		// UpdatedAt has actual timestamp from book's last update
		if existing, exists := userLatest[reading.User]; !exists || reading.UpdatedAt > existing.UpdatedAt {
			userLatest[reading.User] = reading
		}
	}

	// Convert map to list of UserLocation
	users := make([]types.UserLocation, 0, len(userLatest))
	for userName, item := range userLatest {
		users = append(users, types.UserLocation{
			User:      userName,
			AvatarURL: item.ImagemURL,
			CapaURL:   item.CapaURL,
			ISO3:      item.ISO3,
			Pais:      item.Pais,
			Livro:     item.Livro,
			Timestamp: item.UpdatedAt, // v1.0.3: use UpdatedAt instead of SK
		})
	}

	// Build response
	response := types.UserLocationsResponse{
		Users: users,
		Total: len(users),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return errorResponse(500, "Error building response"), nil
	}

	log.Printf("Returning %d unique user locations", len(users))

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(responseBody),
	}, nil
}

func errorResponse(statusCode int, message string) events.APIGatewayV2HTTPResponse {
	body, err := json.Marshal(map[string]string{
		"error": message,
	})
	if err != nil {
		log.Printf("ERROR marshaling error response: %v", err)
		// Fallback to hardcoded JSON if marshal fails
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error":"INTERNAL_ERROR"}`,
		}
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}
}

func main() {
	lambda.Start(handler)
}
