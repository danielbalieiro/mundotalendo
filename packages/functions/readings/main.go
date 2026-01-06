package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	sharedTypes "github.com/mundotalendo/functions/types"
)

// ReadingResponse - Response structure optimized for frontend consumption
type ReadingResponse struct {
	User      string `json:"user"`
	AvatarURL string `json:"avatarURL"`
	CapaURL   string `json:"capaURL"`
	Livro     string `json:"livro"`
	Progresso int    `json:"progresso"`
	Categoria string `json:"categoria"`
	UpdatedAt string `json:"updatedAt"`
}

// Response - API response with all readings
type Response struct {
	Readings []ReadingResponse `json:"readings"`
	Total    int               `json:"total"`
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Extract ISO3 from path parameter
	iso3 := strings.ToUpper(strings.TrimSpace(request.PathParameters["iso3"]))

	// Validate ISO3 format (exactly 3 letters)
	if len(iso3) != 3 || !isAlpha(iso3) {
		return errorResponse(http.StatusBadRequest, "Invalid ISO3 code format"), nil
	}

	// Get table name from environment
	tableName := os.Getenv("SST_Resource_DataTable_name")
	if tableName == "" {
		return errorResponse(http.StatusInternalServerError, "Table name not configured"), nil
	}

	// Initialize DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		return errorResponse(http.StatusInternalServerError, "Failed to load AWS config"), nil
	}
	client := dynamodb.NewFromConfig(cfg)

	// Query DynamoDB for all readings in this country
	readings, err := fetchReadings(ctx, client, tableName, iso3)
	if err != nil {
		return errorResponse(http.StatusInternalServerError, fmt.Sprintf("Database query failed: %v", err)), nil
	}

	// Transform and sort readings
	response := buildResponse(readings)

	// Return JSON response
	body, _ := json.Marshal(response)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func fetchReadings(ctx context.Context, client *dynamodb.Client, tableName, iso3 string) ([]sharedTypes.LeituraItem, error) {
	// Query: PK = "EVENT#LEITURA" with filter on iso3 and progresso >= 1
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		FilterExpression:       aws.String("iso3 = :iso3 AND progresso >= :minProgress"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":          &types.AttributeValueMemberS{Value: "EVENT#LEITURA"},
			":iso3":        &types.AttributeValueMemberS{Value: iso3},
			":minProgress": &types.AttributeValueMemberN{Value: "1"},
		},
	}

	result, err := client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var readings []sharedTypes.LeituraItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &readings)
	if err != nil {
		return nil, err
	}

	return readings, nil
}

func buildResponse(readings []sharedTypes.LeituraItem) Response {
	// Transform to response format
	responses := make([]ReadingResponse, 0, len(readings))
	for _, r := range readings {
		responses = append(responses, ReadingResponse{
			User:      r.User,
			AvatarURL: r.ImagemURL,
			CapaURL:   r.CapaURL,
			Livro:     r.Livro,
			Progresso: r.Progresso,
			Categoria: r.Categoria,
			UpdatedAt: r.UpdatedAt,
		})
	}

	// Sort by: 1) Progresso DESC (completed first), 2) UpdatedAt DESC (newest first)
	sort.Slice(responses, func(i, j int) bool {
		if responses[i].Progresso != responses[j].Progresso {
			return responses[i].Progresso > responses[j].Progresso
		}
		return responses[i].UpdatedAt > responses[j].UpdatedAt
	})

	return Response{
		Readings: responses,
		Total:    len(responses),
	}
}

func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
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
