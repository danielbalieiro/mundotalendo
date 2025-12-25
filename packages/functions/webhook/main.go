package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/mundotalendo/functions/auth"
	"github.com/mundotalendo/functions/mapping"
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

// saveWebhookPayload salva o payload original UMA VEZ por execução de webhook
func saveWebhookPayload(ctx context.Context, webhookUUID, user, payloadBody string, timestamp time.Time) error {
	item := types.WebhookItem{
		PK:      fmt.Sprintf("WEBHOOK#PAYLOAD#%s", webhookUUID),
		SK:      fmt.Sprintf("TIMESTAMP#%s", timestamp.Format(time.RFC3339)),
		User:    user,
		Payload: payloadBody,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook item: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to save webhook payload: %w", err)
	}

	log.Printf("Saved webhook payload for user %s with UUID %s", user, webhookUUID)
	return nil
}

// deleteOldUserReadings deleta todos os registros antigos de leitura do usuário via GSI
func deleteOldUserReadings(ctx context.Context, user string) error {
	// Query usando GSI UserIndex para encontrar todos os registros do usuário
	input := &dynamodb.QueryInput{
		TableName:              &tableName,
		IndexName:              strPtr("UserIndex"),
		KeyConditionExpression: strPtr("#user = :user"),
		ExpressionAttributeNames: map[string]string{
			"#user": "user",
		},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":user": &ddbtypes.AttributeValueMemberS{Value: user},
		},
	}

	result, err := dynamoClient.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query old readings: %w", err)
	}

	if len(result.Items) == 0 {
		log.Printf("No old readings found for user %s", user)
		return nil
	}

	// Deletar cada item encontrado (apenas EVENT#LEITURA, não WEBHOOK#PAYLOAD)
	deletedCount := 0
	for _, item := range result.Items {
		pkAttr, okPK := item["PK"].(*ddbtypes.AttributeValueMemberS)
		skAttr, okSK := item["SK"].(*ddbtypes.AttributeValueMemberS)

		if !okPK || !okSK {
			log.Printf("WARN: Invalid item structure, skipping deletion")
			continue
		}

		// Skip deletion if PK is not an EVENT#LEITURA (protects WEBHOOK#PAYLOAD from being deleted)
		if !strings.HasPrefix(pkAttr.Value, "EVENT#LEITURA") {
			continue
		}

		_, err := dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key: map[string]ddbtypes.AttributeValue{
				"PK": &ddbtypes.AttributeValueMemberS{Value: pkAttr.Value},
				"SK": &ddbtypes.AttributeValueMemberS{Value: skAttr.Value},
			},
		})
		if err != nil {
			log.Printf("WARN: Failed to delete item %s#%s: %v", pkAttr.Value, skAttr.Value, err)
		} else {
			deletedCount++
		}
	}

	log.Printf("Deleted %d old readings for user %s", deletedCount, user)
	return nil
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Printf("Received webhook request: %s", request.Body)

	// Validate payload size (max 1 MB)
	if len(request.Body) > 1024*1024 {
		log.Printf("Payload too large: %d bytes", len(request.Body))
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"PAYLOAD_TOO_LARGE","message":"Payload exceeds 1 MB limit"}`,
		}, nil
	}

	// Validate API key
	apiKey := request.Headers["x-api-key"]
	if apiKey == "" {
		// Try lowercase header
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

	// Parse the webhook payload
	var payload types.WebhookPayload
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		log.Printf("Error parsing payload: %v", err)
		saveToFalhas(ctx, "UNMARSHAL_ERROR", fmt.Sprintf("Failed to parse JSON: %v", err), request.Body)
		// Return 400 for invalid JSON - client error
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"INVALID_JSON","message":"Failed to parse JSON payload"}`,
		}, nil
	}

	// Validate identificador - ignore events from other challenges
	validIdentifiers := []string{"maratona-lendo-paises", "mundotalendo-2026"}
	isValid := false
	for _, id := range validIdentifiers {
		if payload.Maratona.Identificador == id {
			isValid = true
			break
		}
	}
	if !isValid {
		log.Printf("Ignoring event with identificador: %s", payload.Maratona.Identificador)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"success": true, "message": "Event ignored"}`,
		}, nil
	}

	// Validate required fields
	if payload.Perfil.Nome == "" {
		log.Printf("Validation error: missing perfil.nome")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"VALIDATION_ERROR","message":"Missing required field: perfil.nome"}`,
		}, nil
	}

	if len(payload.Desafios) == 0 {
		log.Printf("Validation error: no desafios provided")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"VALIDATION_ERROR","message":"No desafios provided"}`,
		}, nil
	}

	// Generate UUID UMA VEZ para este webhook
	webhookUUID := uuid.New().String()
	user := payload.Perfil.Nome
	timestamp := time.Now()

	log.Printf("Processing webhook %s for user %s", webhookUUID, user)

	// Salvar payload original UMA VEZ
	if err := saveWebhookPayload(ctx, webhookUUID, user, request.Body, timestamp); err != nil {
		log.Printf("WARN: Failed to save webhook payload: %v", err)
		// Não falhar o webhook por isso, apenas logar
	}

	// Deletar leituras antigas do usuário ANTES de criar novas
	if err := deleteOldUserReadings(ctx, user); err != nil {
		log.Printf("WARN: Failed to delete old readings for user %s: %v", user, err)
		// Não falhar o webhook por isso, apenas logar
	}

	type ErrorDetail struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	}

	processedCount := 0
	errorCount := 0
	var errors []ErrorDetail

	// Process each desafio
	for i, desafio := range payload.Desafios {
		// Filter: only process "leitura" or "atividade" types
		if desafio.Tipo != "leitura" && desafio.Tipo != "atividade" {
			continue
		}

		// Calculate max progress from vinculados and extract book title
		maxProgress := 0
		var latestUpdate time.Time
		bookTitle := ""
		for _, vinculado := range desafio.Vinculados {
			if vinculado.Progresso > maxProgress {
				maxProgress = vinculado.Progresso
			}
			// Extract book title from most recent vinculado
			if vinculado.Edicao != nil && vinculado.Edicao.Titulo != "" {
				bookTitle = vinculado.Edicao.Titulo
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

		// Validate progress is in range 0-100
		if maxProgress < 0 || maxProgress > 100 {
			log.Printf("WARN: Invalid progress %d for %s, clamping to 0-100", maxProgress, desafio.Descricao)
			if maxProgress < 0 {
				maxProgress = 0
			} else if maxProgress > 100 {
				maxProgress = 100
			}
		}

		// Clean emojis from country name and category before processing
		cleanedCountryName := utils.CleanEmojis(desafio.Descricao)
		cleanedCategoria := utils.CleanEmojis(desafio.Categoria)

		// Convert country name to ISO code
		iso3 := mapping.GetISO(cleanedCountryName)
		if iso3 == "" {
			log.Printf("Country not found: %s (original: %s)", cleanedCountryName, desafio.Descricao)
			saveToFalhas(ctx, "COUNTRY_NOT_FOUND", "Country not mapped in ISO table", request.Body)
			errors = append(errors, ErrorDetail{
				Code:    "COUNTRY_NOT_FOUND",
				Message: "Country not mapped in ISO table",
				Details: fmt.Sprintf("%s (cleaned from: %s)", cleanedCountryName, desafio.Descricao),
			})
			errorCount++
			continue
		}

		// Create DynamoDB item (v1.0.3: PK simples + SK único + UUID field)
		// PK simples permite queries stats/users funcionarem
		// SK com UUID+ISO+index garante unicidade (múltiplos livros por país)
		// Payload salvo separadamente em WebhookItem (zero duplicação!)
		item := types.LeituraItem{
			PK:          "EVENT#LEITURA",                                // PK simples para queries
			SK:          fmt.Sprintf("%s#%s#%d", webhookUUID, iso3, i), // SK único com índice
			ISO3:        iso3,
			Pais:        cleanedCountryName,
			Categoria:   cleanedCategoria,
			Progresso:   maxProgress,
			User:        user,
			ImagemURL:   payload.Perfil.Imagem,
			Livro:       bookTitle,
			WebhookUUID: webhookUUID,                       // UUID separado para rastreamento
			UpdatedAt:   latestUpdate.Format(time.RFC3339), // Timestamp do último update
		}

		// Save to DynamoDB
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Printf("Error marshaling item: %v", err)
			errors = append(errors, ErrorDetail{
				Code:    "DYNAMODB_MARSHAL_ERROR",
				Message: "Failed to marshal DynamoDB item",
				Details: fmt.Sprintf("%s: %v", desafio.Descricao, err),
			})
			errorCount++
			continue
		}

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      av,
		})
		if err != nil {
			log.Printf("Error saving to DynamoDB: %v", err)
			errors = append(errors, ErrorDetail{
				Code:    "DYNAMODB_PUT_ERROR",
				Message: "Failed to save item to DynamoDB",
				Details: fmt.Sprintf("%s: %v", desafio.Descricao, err),
			})
			errorCount++
			continue
		}

		log.Printf("Processed: %s (%s) - User: %s - Category: %s", cleanedCountryName, iso3, payload.Perfil.Nome, cleanedCategoria)
		processedCount++
	}

	// Build response
	response := map[string]interface{}{
		"success":   processedCount > 0,
		"processed": processedCount,
		"failed":    errorCount,
		"total":     processedCount + errorCount,
	}

	// Add status message
	if processedCount > 0 && errorCount == 0 {
		response["status"] = "COMPLETED"
	} else if processedCount > 0 && errorCount > 0 {
		response["status"] = "PARTIAL"
	} else if errorCount > 0 {
		response["status"] = "FAILED"
	} else {
		response["status"] = "NO_DATA"
	}

	// Add errors array if there are any
	if len(errors) > 0 {
		response["errors"] = errors
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("ERROR marshaling response: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"INTERNAL_ERROR","message":"Failed to build response"}`,
		}, nil
	}

	// Return appropriate status code based on processing results
	statusCode := 200
	if processedCount == 0 && errorCount > 0 {
		// All processing failed - return 500
		statusCode = 500
		log.Printf("All processing failed: %d errors, 0 processed", errorCount)
	} else if processedCount > 0 {
		// At least some items processed successfully - return 200
		statusCode = 200
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
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
				"Content-Type": "application/json",
			},
			Body: `{"error":"INTERNAL_ERROR"}`,
		}
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func saveToFalhas(ctx context.Context, errorType, errorMessage, originalPayload string) {
	if tableName == "" {
		log.Printf("ERROR: tableName is empty - cannot save failure")
		return
	}

	timestamp := time.Now()
	errorUUID := uuid.New().String() // Generate UUID for error tracking
	item := types.FalhaItem{
		PK:              fmt.Sprintf("ERROR#%s", errorUUID),
		SK:              fmt.Sprintf("TIMESTAMP#%s", timestamp.Format(time.RFC3339)),
		ErrorType:       errorType,
		ErrorMessage:    errorMessage,
		OriginalPayload: originalPayload,
	}

	log.Printf("Saving failure to table: %s (ErrorType: %s)", tableName, errorType)

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		log.Printf("ERROR marshaling failure item: %v", err)
		return
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	if err != nil {
		log.Printf("ERROR saving failure: %v", err)
	} else {
		log.Printf("Successfully saved failure")
	}
}

func main() {
	lambda.Start(handler)
}
