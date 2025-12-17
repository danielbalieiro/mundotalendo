package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mundotalendo/functions/mapping"
	"github.com/mundotalendo/functions/types"
)

var (
	dynamoClient    *dynamodb.Client
	tableName       string
	falhasTableName string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	tableName = os.Getenv("SST_Resource_Leituras_name")
	falhasTableName = os.Getenv("SST_Resource_Falhas_name")
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Received webhook request: %s", request.Body)

	// Parse the webhook payload
	var payload types.WebhookPayload
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		log.Printf("Error parsing payload: %v", err)
		saveToFalhas(ctx, "UNMARSHAL_ERROR", fmt.Sprintf("Failed to parse JSON: %v", err), request.Body)
		return errorResponse(400, "Invalid JSON payload"), nil
	}

	// Validate identificador - ignore events from other challenges
	if payload.Maratona.Identificador != "maratona-lendo-paises" {
		log.Printf("Ignoring event with identificador: %s", payload.Maratona.Identificador)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"success": true, "message": "Event ignored"}`,
		}, nil
	}

	processedCount := 0
	errorCount := 0

	// Process each desafio
	for _, desafio := range payload.Desafios {
		// Filter: only process "leitura" or "atividade" types
		if desafio.Tipo != "leitura" && desafio.Tipo != "atividade" {
			continue
		}

		// Calculate max progress from vinculados
		maxProgress := 0
		var latestUpdate time.Time
		for _, vinculado := range desafio.Vinculados {
			if vinculado.Progresso > maxProgress {
				maxProgress = vinculado.Progresso
			}
			// Parse UpdatedAt string to time.Time
			if vinculado.UpdatedAt != "" {
				parsedTime, err := time.Parse("2006-01-02", vinculado.UpdatedAt)
				if err != nil {
					// Try RFC3339 format
					parsedTime, err = time.Parse(time.RFC3339, vinculado.UpdatedAt)
				}
				if err == nil && parsedTime.After(latestUpdate) {
					latestUpdate = parsedTime
				}
			}
		}

		// If no vinculados, use defaults
		if len(desafio.Vinculados) == 0 {
			maxProgress = 0
			latestUpdate = time.Now()
		}

		// If concluido == true, force progress to 100%
		if desafio.Concluido {
			maxProgress = 100
		}

		// Convert country name to ISO code
		iso3 := mapping.GetISO(desafio.Descricao)
		if iso3 == "" {
			log.Printf("Country not found: %s", desafio.Descricao)
			saveToFalhas(ctx, "COUNTRY_NOT_MAPPED",
				fmt.Sprintf("Country name not found: %s", desafio.Descricao),
				request.Body)
			errorCount++
			continue
		}

		// Marshal metadata (complete payload as JSON)
		metadataBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshaling metadata: %v", err)
			saveToFalhas(ctx, "METADATA_MARSHAL_ERROR", err.Error(), request.Body)
			errorCount++
			continue
		}

		// Create DynamoDB item
		timestamp := latestUpdate
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		item := types.LeituraItem{
			PK:        "EVENT#LEITURA",
			SK:        fmt.Sprintf("TIMESTAMP#%s", timestamp.Format(time.RFC3339)),
			ISO3:      iso3,
			Pais:      desafio.Descricao,
			Categoria: desafio.Categoria,
			Progresso: maxProgress,
			User:      payload.Perfil.Nome,
			Metadata:  string(metadataBytes),
		}

		// Save to DynamoDB
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Printf("Error marshaling item: %v", err)
			errorCount++
			continue
		}

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      av,
		})
		if err != nil {
			log.Printf("Error saving to DynamoDB: %v", err)
			errorCount++
			continue
		}

		log.Printf("Processed: %s (%s) - User: %s - Category: %s", desafio.Descricao, iso3, payload.Perfil.Nome, desafio.Categoria)
		processedCount++
	}

	// Return response
	response := map[string]interface{}{
		"success":        true,
		"processedCount": processedCount,
		"errorCount":     errorCount,
		"message":        fmt.Sprintf("Processed %d reading challenges", processedCount),
	}

	responseBody, _ := json.Marshal(response)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
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
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func saveToFalhas(ctx context.Context, errorType, errorMessage, originalPayload string) {
	timestamp := time.Now()
	item := types.FalhaItem{
		PK:              fmt.Sprintf("ERROR#%s", errorType),
		SK:              fmt.Sprintf("TIMESTAMP#%s", timestamp.Format(time.RFC3339)),
		ErrorType:       errorType,
		ErrorMessage:    errorMessage,
		OriginalPayload: originalPayload,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		log.Printf("Error marshaling failure item: %v", err)
		return
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &falhasTableName,
		Item:      av,
	})
	if err != nil {
		log.Printf("Error saving to Falhas table: %v", err)
	}
}

func main() {
	lambda.Start(handler)
}
