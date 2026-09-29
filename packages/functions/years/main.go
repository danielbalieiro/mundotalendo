package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

// handler returns the list of years that have at least one reading registered.
func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Validate API key
	apiKey := request.Headers["x-api-key"]
	if apiKey == "" {
		apiKey = request.Headers["X-API-Key"]
	}
	if !auth.ValidateAPIKey(ctx, dynamoClient, apiKey) {
		return errorResponse(401, "Invalid or missing API key"), nil
	}

	years, err := listYears(ctx)
	if err != nil {
		log.Printf("Error listing years: %v", err)
		return errorResponse(500, "Error fetching years"), nil
	}

	body, err := json.Marshal(map[string]interface{}{
		"years": years,
	})
	if err != nil {
		return errorResponse(500, "Error building response"), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

// listYears scans all reading items and returns the distinct years, sorted
// descending. The year is read from the "year" attribute, falling back to the
// partition key ("EVENT#LEITURA#<year>") for legacy items.
func listYears(ctx context.Context) ([]string, error) {
	yearsSet := make(map[string]bool)

	var lastKey map[string]ddbtypes.AttributeValue
	for {
		result, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
			TableName:            aws.String(tableName),
			FilterExpression:     aws.String("begins_with(PK, :prefix)"),
			ProjectionExpression: aws.String("PK, #y"),
			ExpressionAttributeNames: map[string]string{
				"#y": "year",
			},
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":prefix": &ddbtypes.AttributeValueMemberS{Value: "EVENT#LEITURA"},
			},
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			if y := extractYear(item); y != "" {
				yearsSet[y] = true
			}
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	years := make([]string, 0, len(yearsSet))
	for y := range yearsSet {
		years = append(years, y)
	}

	sort.Slice(years, func(i, j int) bool {
		ni, ei := strconv.Atoi(years[i])
		nj, ej := strconv.Atoi(years[j])
		if ei != nil || ej != nil {
			return years[i] > years[j]
		}
		return ni > nj
	})

	return years, nil
}

// extractYear returns the year from a DynamoDB item, preferring the "year"
// attribute and falling back to the partition key prefix.
func extractYear(item map[string]ddbtypes.AttributeValue) string {
	if attr, ok := item["year"]; ok {
		if s, ok := attr.(*ddbtypes.AttributeValueMemberS); ok && s.Value != "" {
			return s.Value
		}
	}

	pk, ok := item["PK"].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return ""
	}

	rest := strings.TrimPrefix(pk.Value, "EVENT#LEITURA#")
	if rest == pk.Value || rest == "" {
		return ""
	}

	return strings.Split(rest, "#")[0]
}

func errorResponse(statusCode int, message string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
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
