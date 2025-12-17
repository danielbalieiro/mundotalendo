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

func clearTable(ctx context.Context, tableName, pk string) (int, error) {
	result, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]ddbTypes.AttributeValue{
			":pk": &ddbTypes.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range result.Items {
		var keys struct {
			PK string `dynamodbav:"PK"`
			SK string `dynamodbav:"SK"`
		}
		if err := attributevalue.UnmarshalMap(item, &keys); err != nil {
			continue
		}

		keyMap, _ := attributevalue.MarshalMap(map[string]string{
			"PK": keys.PK,
			"SK": keys.SK,
		})
		_, err := dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key:       keyMap,
		})
		if err != nil {
			log.Printf("Error deleting item: %v", err)
			continue
		}
		count++
	}
	return count, nil
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Clearing all data from table")

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
				"Content-Type": "application/json",
			},
			Body: `{"error":"UNAUTHORIZED","message":"Invalid or missing API key"}`,
		}, nil
	}

	eventsDeleted, err := clearTable(ctx, tableName, "EVENT#LEITURA")
	if err != nil {
		log.Printf("Error clearing events: %v", err)
	}

	errorsDeleted := 0
	errorTypes := []string{"COUNTRY_NOT_FOUND", "METADATA_MARSHAL_ERROR", "DYNAMODB_MARSHAL_ERROR", "DYNAMODB_PUT_ERROR"}
	for _, errorType := range errorTypes {
		count, _ := clearTable(ctx, tableName, "ERROR#"+errorType)
		errorsDeleted += count
	}

	response := map[string]interface{}{
		"success":        true,
		"eventsDeleted":  eventsDeleted,
		"errorsDeleted":  errorsDeleted,
		"totalDeleted":   eventsDeleted + errorsDeleted,
	}

	responseBody, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
