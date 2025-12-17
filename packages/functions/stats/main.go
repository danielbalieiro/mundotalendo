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
	log.Println("Fetching stats from DynamoDB")

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

	// Query DynamoDB for all readings
	result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":pk": &ddbTypes.AttributeValueMemberS{Value: "EVENT#LEITURA"},
		},
	})
	if err != nil {
		log.Printf("Error querying DynamoDB: %v", err)
		return errorResponse(500, "Error fetching data"), nil
	}

	// Aggregate max progress per country
	countryProgress := make(map[string]int) // ISO -> max progress
	for _, item := range result.Items {
		var reading types.LeituraItem
		err := attributevalue.UnmarshalMap(item, &reading)
		if err != nil {
			log.Printf("Error unmarshaling item: %v", err)
			continue
		}
		if reading.ISO3 != "" {
			// Keep the maximum progress for each country
			if currentProgress, exists := countryProgress[reading.ISO3]; !exists || reading.Progresso > currentProgress {
				countryProgress[reading.ISO3] = reading.Progresso
			}
		}
	}

	// Convert map to list of CountryProgress
	countries := make([]types.CountryProgress, 0, len(countryProgress))
	for iso, progress := range countryProgress {
		countries = append(countries, types.CountryProgress{
			ISO3:     iso,
			Progress: progress,
		})
	}

	// Build response
	response := types.StatsResponse{
		Countries: countries,
		Total:     len(countries),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return errorResponse(500, "Error building response"), nil
	}

	log.Printf("Returning %d unique countries", len(countries))

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
	body, _ := json.Marshal(map[string]string{
		"error": message,
	})
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
