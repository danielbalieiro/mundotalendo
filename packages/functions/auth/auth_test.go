package auth

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// MockDynamoDBClient implements DynamoDBScanAPI for testing
type MockDynamoDBClient struct {
	ScanFunc func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *MockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if m.ScanFunc != nil {
		return m.ScanFunc(ctx, params, optFns...)
	}
	return &dynamodb.ScanOutput{}, nil
}

func TestValidateAPIKey_EmptyKey(t *testing.T) {
	// Set required environment variable
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	mockClient := &MockDynamoDBClient{}

	result := ValidateAPIKey(context.Background(), mockClient, "")

	if result {
		t.Error("Expected false for empty API key, got true")
	}
}

func TestValidateAPIKey_NoTableName(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("SST_Resource_DataTable_name")

	mockClient := &MockDynamoDBClient{}

	result := ValidateAPIKey(context.Background(), mockClient, "test-key")

	if result {
		t.Error("Expected false when table name is not set, got true")
	}
}

func TestValidateAPIKey_ValidKey(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	validAPIKey := "valid-test-key-123"

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Create a mock API key item
			item := APIKeyItem{
				PK:        "APIKEY#1",
				SK:        "METADATA",
				Name:      "Test API Key",
				Key:       validAPIKey,
				CreatedAt: "2024-12-16T00:00:00Z",
				Active:    true,
			}

			itemMap, err := attributevalue.MarshalMap(item)
			if err != nil {
				t.Fatalf("Failed to marshal item: %v", err)
			}

			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{itemMap},
			}, nil
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, validAPIKey)

	if !result {
		t.Error("Expected true for valid API key, got false")
	}
}

func TestValidateAPIKey_InvalidKey(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Create a different API key
			item := APIKeyItem{
				PK:        "APIKEY#1",
				SK:        "METADATA",
				Name:      "Test API Key",
				Key:       "different-key",
				CreatedAt: "2024-12-16T00:00:00Z",
				Active:    true,
			}

			itemMap, err := attributevalue.MarshalMap(item)
			if err != nil {
				t.Fatalf("Failed to marshal item: %v", err)
			}

			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{itemMap},
			}, nil
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, "wrong-key")

	if result {
		t.Error("Expected false for invalid API key, got true")
	}
}

func TestValidateAPIKey_InactiveKey(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	testAPIKey := "inactive-key-123"

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Create an inactive API key
			item := APIKeyItem{
				PK:        "APIKEY#1",
				SK:        "METADATA",
				Name:      "Inactive Test Key",
				Key:       testAPIKey,
				CreatedAt: "2024-12-16T00:00:00Z",
				Active:    false, // Key is inactive
			}

			itemMap, err := attributevalue.MarshalMap(item)
			if err != nil {
				t.Fatalf("Failed to marshal item: %v", err)
			}

			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{itemMap},
			}, nil
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, testAPIKey)

	if result {
		t.Error("Expected false for inactive API key, got true")
	}
}

func TestValidateAPIKey_NoResults(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Return empty results
			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{},
			}, nil
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, "any-key")

	if result {
		t.Error("Expected false when no API keys found, got true")
	}
}

func TestValidateAPIKey_DynamoDBError(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return nil, errors.New("DynamoDB connection error")
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, "test-key")

	if result {
		t.Error("Expected false when DynamoDB returns error, got true")
	}
}

func TestValidateAPIKey_MultipleKeys(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	validKey := "valid-key-456"

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Create multiple API keys
			items := []APIKeyItem{
				{
					PK:        "APIKEY#1",
					SK:        "METADATA",
					Name:      "First Key",
					Key:       "key-1",
					Active:    true,
				},
				{
					PK:        "APIKEY#2",
					SK:        "METADATA",
					Name:      "Second Key",
					Key:       validKey,
					Active:    true,
				},
				{
					PK:        "APIKEY#3",
					SK:        "METADATA",
					Name:      "Third Key",
					Key:       "key-3",
					Active:    true,
				},
			}

			var itemMaps []map[string]ddbTypes.AttributeValue
			for _, item := range items {
				itemMap, err := attributevalue.MarshalMap(item)
				if err != nil {
					t.Fatalf("Failed to marshal item: %v", err)
				}
				itemMaps = append(itemMaps, itemMap)
			}

			return &dynamodb.ScanOutput{
				Items: itemMaps,
			}, nil
		},
	}

	result := ValidateAPIKey(context.Background(), mockClient, validKey)

	if !result {
		t.Error("Expected true when valid key is among multiple keys, got false")
	}
}

func TestValidateAPIKey_ScanParameters(t *testing.T) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	var capturedParams *dynamodb.ScanInput

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			capturedParams = params
			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{},
			}, nil
		},
	}

	ValidateAPIKey(context.Background(), mockClient, "test-key")

	if capturedParams == nil {
		t.Fatal("Scan was not called")
	}

	// Verify table name
	if *capturedParams.TableName != "test-table" {
		t.Errorf("Expected table name 'test-table', got '%s'", *capturedParams.TableName)
	}

	// Verify filter expression
	if capturedParams.FilterExpression == nil {
		t.Error("Expected FilterExpression to be set")
	}

	// Verify expression attribute names
	if capturedParams.ExpressionAttributeNames == nil {
		t.Error("Expected ExpressionAttributeNames to be set")
	}

	// Verify expression attribute values
	if capturedParams.ExpressionAttributeValues == nil {
		t.Error("Expected ExpressionAttributeValues to be set")
	}
}

func TestAPIKeyItem_Structure(t *testing.T) {
	item := APIKeyItem{
		PK:        "APIKEY#test",
		SK:        "METADATA",
		Name:      "Test Key",
		Key:       "secret-key-123",
		CreatedAt: "2024-12-16T00:00:00Z",
		Active:    true,
	}

	// Test that marshaling and unmarshaling preserves data
	itemMap, err := attributevalue.MarshalMap(item)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled APIKeyItem
	err = attributevalue.UnmarshalMap(itemMap, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshaled.PK != item.PK {
		t.Errorf("PK mismatch: got %s, want %s", unmarshaled.PK, item.PK)
	}
	if unmarshaled.SK != item.SK {
		t.Errorf("SK mismatch: got %s, want %s", unmarshaled.SK, item.SK)
	}
	if unmarshaled.Name != item.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, item.Name)
	}
	if unmarshaled.Key != item.Key {
		t.Errorf("Key mismatch: got %s, want %s", unmarshaled.Key, item.Key)
	}
	if unmarshaled.Active != item.Active {
		t.Errorf("Active mismatch: got %v, want %v", unmarshaled.Active, item.Active)
	}
}

func BenchmarkValidateAPIKey_ValidKey(b *testing.B) {
	os.Setenv("SST_Resource_DataTable_name", "test-table")
	defer os.Unsetenv("SST_Resource_DataTable_name")

	validAPIKey := "benchmark-key"

	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			item := APIKeyItem{
				PK:     "APIKEY#1",
				SK:     "METADATA",
				Name:   "Benchmark Key",
				Key:    validAPIKey,
				Active: true,
			}

			itemMap, _ := attributevalue.MarshalMap(item)

			return &dynamodb.ScanOutput{
				Items: []map[string]ddbTypes.AttributeValue{itemMap},
			}, nil
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAPIKey(ctx, mockClient, validAPIKey)
	}
}
