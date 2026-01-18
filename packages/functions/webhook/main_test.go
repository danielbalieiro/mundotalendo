package main

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mundotalendo/functions/types"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		code           string
		message        string
		expectedStatus int
	}{
		{
			name:           "400 Bad Request",
			statusCode:     400,
			code:           "INVALID_JSON",
			message:        "Invalid request",
			expectedStatus: 400,
		},
		{
			name:           "401 Unauthorized",
			statusCode:     401,
			code:           "UNAUTHORIZED",
			message:        "Invalid API key",
			expectedStatus: 401,
		},
		{
			name:           "500 Internal Server Error",
			statusCode:     500,
			code:           "STORAGE_ERROR",
			message:        "Failed to store payload",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := errorResponse(tt.statusCode, tt.code, tt.message)

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if response.Headers["Content-Type"] != "application/json" {
				t.Error("Expected Content-Type to be application/json")
			}

			var body map[string]string
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if body["error"] != tt.code {
				t.Errorf("Expected error code '%s', got '%s'", tt.code, body["error"])
			}

			if body["message"] != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, body["message"])
			}
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	response := successResponse("Test message")

	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", response.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if body["success"] != true {
		t.Error("Expected success to be true")
	}

	if body["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got '%v'", body["message"])
	}
}

func TestAcceptedResponse(t *testing.T) {
	uuid := "test-uuid-12345"
	response := acceptedResponse(uuid)

	if response.StatusCode != 202 {
		t.Errorf("Expected status code 202, got %d", response.StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if body["success"] != true {
		t.Error("Expected success to be true")
	}

	if body["uuid"] != uuid {
		t.Errorf("Expected uuid '%s', got '%v'", uuid, body["uuid"])
	}

	if body["status"] != "QUEUED" {
		t.Errorf("Expected status 'QUEUED', got '%v'", body["status"])
	}
}

func TestWebhookPayload_Parsing(t *testing.T) {
	validPayload := `{
		"perfil": {
			"nome": "Test User",
			"link": "https://test.com/user",
			"imagem": "https://example.com/avatar.jpg"
		},
		"maratona": {
			"nome": "Test Marathon",
			"identificador": "maratona-lendo-paises"
		},
		"desafios": [
			{
				"descricao": "Brasil",
				"categoria": "Janeiro",
				"concluido": true,
				"tipo": "leitura",
				"vinculados": [
					{
						"completo": true,
						"progresso": 100,
						"updatedAt": "2024-12-16T00:00:00Z"
					}
				]
			}
		]
	}`

	var payload types.WebhookPayload
	err := json.Unmarshal([]byte(validPayload), &payload)

	if err != nil {
		t.Fatalf("Failed to parse valid payload: %v", err)
	}

	if payload.Perfil.Nome != "Test User" {
		t.Errorf("Expected perfil nome 'Test User', got '%s'", payload.Perfil.Nome)
	}

	if payload.Perfil.Imagem != "https://example.com/avatar.jpg" {
		t.Errorf("Expected perfil imagem URL, got '%s'", payload.Perfil.Imagem)
	}

	if payload.Maratona.Identificador != "maratona-lendo-paises" {
		t.Errorf("Expected identificador 'maratona-lendo-paises', got '%s'", payload.Maratona.Identificador)
	}

	if len(payload.Desafios) != 1 {
		t.Errorf("Expected 1 desafio, got %d", len(payload.Desafios))
	}

	if payload.Desafios[0].Descricao != "Brasil" {
		t.Errorf("Expected descricao 'Brasil', got '%s'", payload.Desafios[0].Descricao)
	}
}

func TestWebhookPayload_InvalidJSON(t *testing.T) {
	invalidPayloads := []string{
		`{invalid json}`,
		``,
		`{"perfil": "invalid"}`,
	}

	for i, invalidPayload := range invalidPayloads {
		var payload types.WebhookPayload
		err := json.Unmarshal([]byte(invalidPayload), &payload)

		if err == nil && invalidPayload != "" {
			t.Errorf("Test case %d: Expected error for invalid JSON, got nil", i)
		}
	}
}

func TestValidIdentifiers(t *testing.T) {
	tests := []struct {
		identificador string
		expected      bool
	}{
		{"maratona-lendo-paises", true},
		{"mundotalendo-2026", true},
		{"other-marathon", false},
		{"", false},
		{"random-id", false},
	}

	for _, tt := range tests {
		t.Run(tt.identificador, func(t *testing.T) {
			result := ValidIdentifiers[tt.identificador]
			if result != tt.expected {
				t.Errorf("ValidIdentifiers[%q] = %v, want %v", tt.identificador, result, tt.expected)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedAPIKey string
	}{
		{
			name: "Lowercase header",
			headers: map[string]string{
				"x-api-key": "test-key-123",
			},
			expectedAPIKey: "test-key-123",
		},
		{
			name: "Capitalized header",
			headers: map[string]string{
				"X-API-Key": "test-key-456",
			},
			expectedAPIKey: "test-key-456",
		},
		{
			name:           "No API key header",
			headers:        map[string]string{},
			expectedAPIKey: "",
		},
		{
			name: "Both headers - lowercase priority",
			headers: map[string]string{
				"x-api-key": "lowercase-key",
				"X-API-Key": "capitalized-key",
			},
			expectedAPIKey: "lowercase-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey := getAPIKey(tt.headers)

			if apiKey != tt.expectedAPIKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.expectedAPIKey, apiKey)
			}
		})
	}
}

func TestAPIGatewayRequest_Headers(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedAPIKey string
	}{
		{
			name: "Lowercase header",
			headers: map[string]string{
				"x-api-key": "test-key-123",
			},
			expectedAPIKey: "test-key-123",
		},
		{
			name: "Capitalized header",
			headers: map[string]string{
				"X-API-Key": "test-key-456",
			},
			expectedAPIKey: "test-key-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := events.APIGatewayV2HTTPRequest{
				Headers: tt.headers,
			}

			apiKey := getAPIKey(request.Headers)

			if apiKey != tt.expectedAPIKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.expectedAPIKey, apiKey)
			}
		})
	}
}

func TestMaxPayloadSize(t *testing.T) {
	if MaxPayloadSize != 1024*1024 {
		t.Errorf("Expected MaxPayloadSize to be 1MB (1048576), got %d", MaxPayloadSize)
	}
}

func TestSQSMessage_Structure(t *testing.T) {
	msg := types.SQSMessage{
		UUID:      "test-uuid-123",
		User:      "Test User",
		Timestamp: "2026-01-14T10:00:00Z",
	}

	if msg.UUID != "test-uuid-123" {
		t.Errorf("Expected UUID 'test-uuid-123', got '%s'", msg.UUID)
	}

	if msg.User != "Test User" {
		t.Errorf("Expected User 'Test User', got '%s'", msg.User)
	}

	if msg.Timestamp != "2026-01-14T10:00:00Z" {
		t.Errorf("Expected Timestamp '2026-01-14T10:00:00Z', got '%s'", msg.Timestamp)
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal SQSMessage: %v", err)
	}

	var unmarshaled types.SQSMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SQSMessage: %v", err)
	}

	if unmarshaled.UUID != msg.UUID {
		t.Error("SQSMessage lost data during marshal/unmarshal")
	}
}

func TestFalhaItem_Structure(t *testing.T) {
	item := types.FalhaItem{
		PK:              "ERROR#COUNTRY_NOT_FOUND",
		SK:              "TIMESTAMP#2024-12-16T00:00:00Z",
		ErrorType:       "COUNTRY_NOT_FOUND",
		ErrorMessage:    "Country not mapped",
		OriginalPayload: `{"test":"payload"}`,
	}

	if item.ErrorType != "COUNTRY_NOT_FOUND" {
		t.Errorf("Expected ErrorType 'COUNTRY_NOT_FOUND', got '%s'", item.ErrorType)
	}

	if item.OriginalPayload == "" {
		t.Error("Expected OriginalPayload to be set")
	}
}

func TestPerfilImagemField(t *testing.T) {
	perfil := types.Perfil{
		Nome:   "Test User",
		Link:   "https://example.com/user",
		Imagem: "https://storage.googleapis.com/example/avatar.jpg",
	}

	if perfil.Imagem == "" {
		t.Error("Expected Perfil.Imagem to be set")
	}

	if perfil.Imagem != "https://storage.googleapis.com/example/avatar.jpg" {
		t.Errorf("Expected Imagem URL to match, got '%s'", perfil.Imagem)
	}
}

func TestLeituraItem_Structure(t *testing.T) {
	item := types.LeituraItem{
		PK:          "EVENT#LEITURA",
		SK:          "test-uuid#BRA#0",
		ISO3:        "BRA",
		Pais:        "Brasil",
		Categoria:   "Janeiro",
		Progresso:   100,
		User:        "Test User",
		ImagemURL:   "https://example.com/avatar.jpg",
		WebhookUUID: "test-uuid",
		UpdatedAt:   "2026-01-14T10:00:00Z",
	}

	if item.PK != "EVENT#LEITURA" {
		t.Errorf("Expected PK 'EVENT#LEITURA', got '%s'", item.PK)
	}

	if item.ISO3 != "BRA" {
		t.Errorf("Expected ISO3 'BRA', got '%s'", item.ISO3)
	}

	if item.Progresso != 100 {
		t.Errorf("Expected Progresso 100, got %d", item.Progresso)
	}

	if item.WebhookUUID != "test-uuid" {
		t.Errorf("Expected WebhookUUID 'test-uuid', got '%s'", item.WebhookUUID)
	}

	if item.ImagemURL != "https://example.com/avatar.jpg" {
		t.Errorf("Expected ImagemURL to be set, got '%s'", item.ImagemURL)
	}
}

func TestDesafioTypeFiltering(t *testing.T) {
	desafios := []types.Desafio{
		{Tipo: "leitura", Descricao: "Brasil"},
		{Tipo: "atividade", Descricao: "Portugal"},
		{Tipo: "outro", Descricao: "França"},
		{Tipo: "desconhecido", Descricao: "Alemanha"},
		{Tipo: "leitura", Descricao: "Japão"},
	}

	validTypes := map[string]bool{
		"leitura":   true,
		"atividade": true,
	}

	validCount := 0
	for _, desafio := range desafios {
		if validTypes[desafio.Tipo] {
			validCount++
		}
	}

	expectedValid := 3 // leitura (2) + atividade (1)
	if validCount != expectedValid {
		t.Errorf("Expected %d valid desafios, got %d", expectedValid, validCount)
	}
}
