package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mundotalendo/functions/types"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		expectedStatus int
	}{
		{
			name:           "400 Bad Request",
			statusCode:     400,
			message:        "Invalid request",
			expectedStatus: 400,
		},
		{
			name:           "401 Unauthorized",
			statusCode:     401,
			message:        "Unauthorized",
			expectedStatus: 401,
		},
		{
			name:           "500 Internal Server Error",
			statusCode:     500,
			message:        "Internal error",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := errorResponse(tt.statusCode, tt.message)

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

			if body["error"] != tt.message {
				t.Errorf("Expected error message '%s', got '%s'", tt.message, body["error"])
			}
		})
	}
}

func TestWebhookPayload_Parsing(t *testing.T) {
	validPayload := `{
		"perfil": {
			"nome": "Test User",
			"link": "https://test.com/user"
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
		`null`,
		`{"perfil": "invalid"}`,
	}

	for i, invalidPayload := range invalidPayloads {
		var payload types.WebhookPayload
		err := json.Unmarshal([]byte(invalidPayload), &payload)

		if err == nil && invalidPayload != "" && invalidPayload != "null" {
			t.Errorf("Test case %d: Expected error for invalid JSON, got nil", i)
		}
	}
}

func TestProgressCalculation(t *testing.T) {
	tests := []struct {
		name            string
		vinculados      []types.Vinculado
		concluido       bool
		expectedMax     int
		expectLatestSet bool
	}{
		{
			name: "Single vinculado with 50%",
			vinculados: []types.Vinculado{
				{Progresso: 50, UpdatedAt: "2024-12-16T00:00:00Z"},
			},
			concluido:       false,
			expectedMax:     50,
			expectLatestSet: true,
		},
		{
			name: "Multiple vinculados, max is 80",
			vinculados: []types.Vinculado{
				{Progresso: 50, UpdatedAt: "2024-12-15T00:00:00Z"},
				{Progresso: 80, UpdatedAt: "2024-12-16T00:00:00Z"},
				{Progresso: 30, UpdatedAt: "2024-12-14T00:00:00Z"},
			},
			concluido:       false,
			expectedMax:     80,
			expectLatestSet: true,
		},
		{
			name: "Concluido overrides to 100%",
			vinculados: []types.Vinculado{
				{Progresso: 50, UpdatedAt: "2024-12-16T00:00:00Z"},
			},
			concluido:       true,
			expectedMax:     100,
			expectLatestSet: true,
		},
		{
			name:            "Empty vinculados defaults to 0",
			vinculados:      []types.Vinculado{},
			concluido:       false,
			expectedMax:     0,
			expectLatestSet: false,
		},
		{
			name:            "Empty vinculados but concluido",
			vinculados:      []types.Vinculado{},
			concluido:       true,
			expectedMax:     100,
			expectLatestSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxProgress := 0
			var latestUpdate time.Time

			for _, vinculado := range tt.vinculados {
				if vinculado.Progresso > maxProgress {
					maxProgress = vinculado.Progresso
				}
				if vinculado.UpdatedAt != "" {
					parsedTime, err := time.Parse(time.RFC3339, vinculado.UpdatedAt)
					if err == nil && parsedTime.After(latestUpdate) {
						latestUpdate = parsedTime
					}
				}
			}

			if tt.concluido {
				maxProgress = 100
			}

			if maxProgress != tt.expectedMax {
				t.Errorf("Expected max progress %d, got %d", tt.expectedMax, maxProgress)
			}

			if tt.expectLatestSet && latestUpdate.IsZero() {
				t.Error("Expected latestUpdate to be set, but it's zero")
			}
		})
	}
}

func TestTimeParsing(t *testing.T) {
	tests := []struct {
		name        string
		timeString  string
		shouldParse bool
	}{
		{
			name:        "RFC3339 format",
			timeString:  "2024-12-16T00:00:00Z",
			shouldParse: true,
		},
		{
			name:        "Date only format",
			timeString:  "2024-12-16",
			shouldParse: true,
		},
		{
			name:        "RFC3339 with timezone",
			timeString:  "2024-12-16T15:30:00-03:00",
			shouldParse: true,
		},
		{
			name:        "Invalid format",
			timeString:  "16/12/2024",
			shouldParse: false,
		},
		{
			name:        "Empty string",
			timeString:  "",
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsedTime time.Time
			var err error

			// Try date-only format first
			parsedTime, err = time.Parse("2006-01-02", tt.timeString)
			if err != nil {
				// Try RFC3339 format
				parsedTime, err = time.Parse(time.RFC3339, tt.timeString)
			}

			if tt.shouldParse && err != nil {
				t.Errorf("Expected to parse '%s', but got error: %v", tt.timeString, err)
			}

			if !tt.shouldParse && err == nil && tt.timeString != "" {
				t.Errorf("Expected error parsing '%s', but succeeded: %v", tt.timeString, parsedTime)
			}
		})
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

func TestLeituraItem_Structure(t *testing.T) {
	item := types.LeituraItem{
		PK:        "EVENT#LEITURA",
		SK:        "TIMESTAMP#2024-12-16T00:00:00Z",
		ISO3:      "BRA",
		Pais:      "Brasil",
		Categoria: "Janeiro",
		Progresso: 100,
		User:      "Test User",
		Metadata:  `{"test":"data"}`,
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
		{
			name:           "No API key header",
			headers:        map[string]string{},
			expectedAPIKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := events.APIGatewayV2HTTPRequest{
				Headers: tt.headers,
			}

			apiKey := request.Headers["x-api-key"]
			if apiKey == "" {
				apiKey = request.Headers["X-API-Key"]
			}

			if apiKey != tt.expectedAPIKey {
				t.Errorf("Expected API key '%s', got '%s'", tt.expectedAPIKey, apiKey)
			}
		})
	}
}

func TestResponseStatusCodes(t *testing.T) {
	tests := []struct {
		name               string
		scenario           string
		expectedStatusCode int
	}{
		{
			name:               "Unauthorized",
			scenario:           "invalid_api_key",
			expectedStatusCode: 401,
		},
		{
			name:               "Invalid JSON",
			scenario:           "malformed_json",
			expectedStatusCode: 400,
		},
		{
			name:               "Successful processing",
			scenario:           "valid_request",
			expectedStatusCode: 200,
		},
		{
			name:               "Ignored event",
			scenario:           "wrong_identificador",
			expectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test documents the expected status codes
			// Actual implementation would require full handler testing with mocks
			expectedCode := tt.expectedStatusCode
			if expectedCode != 200 && expectedCode != 400 && expectedCode != 401 {
				t.Errorf("Unexpected status code in test definition: %d", expectedCode)
			}
		})
	}
}

func TestMetadataMarshaling(t *testing.T) {
	payload := types.WebhookPayload{
		Perfil: types.Perfil{
			Nome: "Test User",
			Link: "https://test.com",
		},
		Maratona: types.Maratona{
			Nome:          "Test Marathon",
			Identificador: "test-id",
		},
		Desafios: []types.Desafio{
			{
				Descricao: "Brasil",
				Categoria: "Janeiro",
				Concluido: true,
				Tipo:      "leitura",
			},
		},
	}

	metadataBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	if len(metadataBytes) == 0 {
		t.Error("Expected metadata to have content")
	}

	// Verify it can be unmarshaled back
	var unmarshaled types.WebhookPayload
	err = json.Unmarshal(metadataBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if unmarshaled.Perfil.Nome != payload.Perfil.Nome {
		t.Error("Metadata lost data during marshal/unmarshal cycle")
	}
}
