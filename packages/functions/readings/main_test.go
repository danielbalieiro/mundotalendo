package main

import (
	"testing"

	sharedTypes "github.com/mundotalendo/functions/types"
)

func TestBuildResponse(t *testing.T) {
	readings := []sharedTypes.LeituraItem{
		{User: "Alice", Livro: "Book A", Progresso: 100, UpdatedAt: "2025-12-20T10:00:00Z"},
		{User: "Bob", Livro: "Book B", Progresso: 50, UpdatedAt: "2025-12-21T10:00:00Z"},
		{User: "Charlie", Livro: "Book C", Progresso: 100, UpdatedAt: "2025-12-19T10:00:00Z"},
	}

	response := buildResponse(readings)

	if response.Total != 3 {
		t.Errorf("Expected total 3, got %d", response.Total)
	}

	// Should be sorted: completed books first, then by date
	if response.Readings[0].User != "Alice" {
		t.Errorf("Expected Alice first (100%%, newest), got %s", response.Readings[0].User)
	}
	if response.Readings[1].User != "Charlie" {
		t.Errorf("Expected Charlie second (100%%, older), got %s", response.Readings[1].User)
	}
	if response.Readings[2].User != "Bob" {
		t.Errorf("Expected Bob third (50%%), got %s", response.Readings[2].User)
	}
}

func TestIsAlpha(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"BRA", true},
		{"USA", true},
		{"BR1", false},
		{"BR", true},
		{"BRAZ", true},
		{"123", false},
		{"", true}, // edge case: empty string
		{"AB-", false},
		{"xyz", true},
	}

	for _, tt := range tests {
		result := isAlpha(tt.input)
		if result != tt.expected {
			t.Errorf("isAlpha(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestBuildResponseEmptyInput(t *testing.T) {
	readings := []sharedTypes.LeituraItem{}
	response := buildResponse(readings)

	if response.Total != 0 {
		t.Errorf("Expected total 0, got %d", response.Total)
	}

	if len(response.Readings) != 0 {
		t.Errorf("Expected empty readings array, got length %d", len(response.Readings))
	}
}

func TestBuildResponseSortingByProgressOnly(t *testing.T) {
	readings := []sharedTypes.LeituraItem{
		{User: "Alice", Livro: "Book A", Progresso: 30, UpdatedAt: "2025-12-20T10:00:00Z"},
		{User: "Bob", Livro: "Book B", Progresso: 80, UpdatedAt: "2025-12-21T10:00:00Z"},
		{User: "Charlie", Livro: "Book C", Progresso: 50, UpdatedAt: "2025-12-19T10:00:00Z"},
	}

	response := buildResponse(readings)

	// Should be sorted by progress DESC
	if response.Readings[0].Progresso != 80 {
		t.Errorf("Expected first reading with 80%%, got %d%%", response.Readings[0].Progresso)
	}
	if response.Readings[1].Progresso != 50 {
		t.Errorf("Expected second reading with 50%%, got %d%%", response.Readings[1].Progresso)
	}
	if response.Readings[2].Progresso != 30 {
		t.Errorf("Expected third reading with 30%%, got %d%%", response.Readings[2].Progresso)
	}
}
