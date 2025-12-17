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
)

var (
	dynamoClient  *dynamodb.Client
	leiturasTable string
	falhasTable   string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	leiturasTable = os.Getenv("SST_Resource_Leituras_name")
	falhasTable = os.Getenv("SST_Resource_Falhas_name")
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
	log.Println("Clearing all tables")

	leiturasDeleted, err := clearTable(ctx, leiturasTable, "EVENT#LEITURA")
	if err != nil {
		log.Printf("Error clearing Leituras: %v", err)
	}

	falhasDeleted := 0
	errorTypes := []string{"COUNTRY_NOT_MAPPED", "METADATA_MARSHAL_ERROR"}
	for _, errorType := range errorTypes {
		count, _ := clearTable(ctx, falhasTable, "ERROR#"+errorType)
		falhasDeleted += count
	}

	response := map[string]interface{}{
		"success":         true,
		"leiturasDeleted": leiturasDeleted,
		"falhasDeleted":   falhasDeleted,
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
