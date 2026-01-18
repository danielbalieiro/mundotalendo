package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mundotalendo/functions/types"
)

// Mock S3 client for testing
type mockS3Client struct {
	payload string
	err     error
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(m.payload)),
	}, nil
}

// Mock DynamoDB client for testing
type mockDynamoDBClient struct {
	putErr    error
	queryErr  error
	deleteErr error
	items     []map[string]interface{}
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, m.putErr
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, m.queryErr
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, m.deleteErr
}

func TestExtractDesafioData(t *testing.T) {
	tests := []struct {
		name         string
		desafio      types.Desafio
		wantProgress int
		wantTitle    string
		wantCapa     string
	}{
		{
			name: "single vinculado with progress",
			desafio: types.Desafio{
				Vinculados: []types.Vinculado{
					{
						Progresso: 50,
						Edicao: &types.Edicao{
							Titulo: "Test Book",
							Capa:   "http://example.com/capa.jpg",
						},
					},
				},
			},
			wantProgress: 50,
			wantTitle:    "Test Book",
			wantCapa:     "http://example.com/capa.jpg",
		},
		{
			name: "multiple vinculados - max progress",
			desafio: types.Desafio{
				Vinculados: []types.Vinculado{
					{Progresso: 25},
					{Progresso: 75},
					{Progresso: 50},
				},
			},
			wantProgress: 75,
		},
		{
			name: "concluido forces 100%",
			desafio: types.Desafio{
				Concluido: true,
				Vinculados: []types.Vinculado{
					{Progresso: 50},
				},
			},
			wantProgress: 100,
		},
		{
			name:         "empty vinculados",
			desafio:      types.Desafio{},
			wantProgress: 0,
		},
		{
			name: "progress over 100 clamped",
			desafio: types.Desafio{
				Vinculados: []types.Vinculado{
					{Progresso: 150},
				},
			},
			wantProgress: 100,
		},
		{
			name: "negative progress clamped",
			desafio: types.Desafio{
				Vinculados: []types.Vinculado{
					{Progresso: -10},
				},
			},
			wantProgress: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress, _, title, capa := extractDesafioData(tt.desafio)

			if progress != tt.wantProgress {
				t.Errorf("progress = %d, want %d", progress, tt.wantProgress)
			}
			if tt.wantTitle != "" && title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if tt.wantCapa != "" && capa != tt.wantCapa {
				t.Errorf("capa = %q, want %q", capa, tt.wantCapa)
			}
		})
	}
}

func TestClampProgress(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: -10, want: 0},
		{input: 0, want: 0},
		{input: 50, want: 50},
		{input: 100, want: 100},
		{input: 150, want: 100},
	}

	for _, tt := range tests {
		got := clampProgress(tt.input)
		if got != tt.want {
			t.Errorf("clampProgress(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "invalid message - not retryable",
			err:  ErrInvalidMessage,
			want: false,
		},
		{
			name: "payload not found - not retryable",
			err:  ErrPayloadNotFound,
			want: false,
		},
		{
			name: "invalid payload - not retryable",
			err:  ErrInvalidPayload,
			want: false,
		},
		{
			name: "S3 fetch error - retryable",
			err:  ErrS3Fetch,
			want: true,
		},
		{
			name: "DynamoDB write error - retryable",
			err:  ErrDynamoDBWrite,
			want: true,
		},
		{
			name: "country not found - retryable (default)",
			err:  ErrCountryNotFound,
			want: true,
		},
		{
			name: "unknown error - retryable (default)",
			err:  errors.New("unknown error"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestPayloadFetcher_FetchPayload(t *testing.T) {
	validPayload := `{
		"perfil": {"nome": "Test User", "imagem": "http://example.com/avatar.jpg"},
		"maratona": {"identificador": "mundotalendo-2026"},
		"desafios": [{"descricao": "Brasil", "tipo": "leitura"}]
	}`

	tests := []struct {
		name    string
		payload string
		err     error
		wantErr bool
	}{
		{
			name:    "valid payload",
			payload: validPayload,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			payload: "not json",
			wantErr: true,
		},
		{
			name:    "S3 error",
			err:     errors.New("S3 error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockS3Client{payload: tt.payload, err: tt.err}
			fetcher := NewPayloadFetcher(client, "test-bucket")

			payload, err := fetcher.FetchPayload(context.Background(), "test-uuid")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if payload.Perfil.Nome != "Test User" {
				t.Errorf("user = %q, want %q", payload.Perfil.Nome, "Test User")
			}
		})
	}
}

func TestProcessingMeta(t *testing.T) {
	meta := ProcessingMeta{
		UUID:      "test-uuid",
		User:      "Test User",
		AvatarURL: "http://example.com/avatar.jpg",
		Timestamp: time.Now(),
	}

	if meta.UUID != "test-uuid" {
		t.Errorf("UUID = %q, want %q", meta.UUID, "test-uuid")
	}
	if meta.User != "Test User" {
		t.Errorf("User = %q, want %q", meta.User, "Test User")
	}
}

func TestValidDesafioTypes(t *testing.T) {
	tests := []struct {
		tipo  string
		valid bool
	}{
		{"leitura", true},
		{"atividade", true},
		{"outro", false},
		{"", false},
	}

	for _, tt := range tests {
		got := ValidDesafioTypes[tt.tipo]
		if got != tt.valid {
			t.Errorf("ValidDesafioTypes[%q] = %v, want %v", tt.tipo, got, tt.valid)
		}
	}
}

func TestWrapError(t *testing.T) {
	err := WrapError("test_op", "test-uuid", "Brasil", ErrDynamoDBWrite)

	if !errors.Is(err, ErrDynamoDBWrite) {
		t.Error("WrapError should preserve underlying error")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "test_op") {
		t.Errorf("error message should contain operation: %s", errMsg)
	}
	if !strings.Contains(errMsg, "test-uuid") {
		t.Errorf("error message should contain UUID: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Brasil") {
		t.Errorf("error message should contain country: %s", errMsg)
	}
}

func TestProcessingResult(t *testing.T) {
	// Success result
	success := ProcessingResult{
		ISO3:      "BRA",
		Country:   "Brasil",
		Processed: true,
	}
	if !success.Processed {
		t.Error("success result should be processed")
	}

	// Error result
	failure := ProcessingResult{
		Country: "Unknown",
		Error:   ErrCountryNotFound,
	}
	if failure.Processed {
		t.Error("failure result should not be processed")
	}
	if failure.Error == nil {
		t.Error("failure result should have error")
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid RFC3339",
			input: "2026-01-14T10:00:00Z",
			valid: true,
		},
		{
			name:  "invalid format",
			input: "not-a-timestamp",
			valid: false,
		},
		{
			name:  "empty string",
			input: "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimestamp(tt.input)

			if tt.valid {
				// Should parse to the expected time
				if result.IsZero() {
					t.Error("expected valid time, got zero")
				}
			} else {
				// Should return current time (not zero, but recent)
				if result.IsZero() {
					t.Error("expected fallback to current time, got zero")
				}
			}
		})
	}
}
