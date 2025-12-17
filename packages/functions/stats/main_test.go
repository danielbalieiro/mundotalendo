package main

import (
	"encoding/json"
	"testing"

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
			name:           "500 Internal Server Error",
			statusCode:     500,
			message:        "Error fetching data",
			expectedStatus: 500,
		},
		{
			name:           "401 Unauthorized",
			statusCode:     401,
			message:        "Unauthorized",
			expectedStatus: 401,
		},
		{
			name:           "400 Bad Request",
			statusCode:     400,
			message:        "Bad request",
			expectedStatus: 400,
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

			if response.Headers["Access-Control-Allow-Origin"] != "*" {
				t.Error("Expected CORS header to be set")
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

func TestCountryProgressAggregation(t *testing.T) {
	tests := []struct {
		name            string
		readings        []types.LeituraItem
		expectedMax     map[string]int
		expectedTotal   int
	}{
		{
			name: "Single country single reading",
			readings: []types.LeituraItem{
				{ISO3: "BRA", Progresso: 50},
			},
			expectedMax: map[string]int{
				"BRA": 50,
			},
			expectedTotal: 1,
		},
		{
			name: "Single country multiple readings - max progress",
			readings: []types.LeituraItem{
				{ISO3: "BRA", Progresso: 30},
				{ISO3: "BRA", Progresso: 80},
				{ISO3: "BRA", Progresso: 50},
			},
			expectedMax: map[string]int{
				"BRA": 80,
			},
			expectedTotal: 1,
		},
		{
			name: "Multiple countries",
			readings: []types.LeituraItem{
				{ISO3: "BRA", Progresso: 100},
				{ISO3: "USA", Progresso: 75},
				{ISO3: "JPN", Progresso: 50},
			},
			expectedMax: map[string]int{
				"BRA": 100,
				"USA": 75,
				"JPN": 50,
			},
			expectedTotal: 3,
		},
		{
			name: "Multiple countries with multiple readings",
			readings: []types.LeituraItem{
				{ISO3: "BRA", Progresso: 30},
				{ISO3: "BRA", Progresso: 100},
				{ISO3: "USA", Progresso: 50},
				{ISO3: "USA", Progresso: 75},
				{ISO3: "JPN", Progresso: 60},
			},
			expectedMax: map[string]int{
				"BRA": 100,
				"USA": 75,
				"JPN": 60,
			},
			expectedTotal: 3,
		},
		{
			name:          "Empty readings",
			readings:      []types.LeituraItem{},
			expectedMax:   map[string]int{},
			expectedTotal: 0,
		},
		{
			name: "Reading with empty ISO3",
			readings: []types.LeituraItem{
				{ISO3: "", Progresso: 50},
				{ISO3: "BRA", Progresso: 100},
			},
			expectedMax: map[string]int{
				"BRA": 100,
			},
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the aggregation logic from handler
			countryProgress := make(map[string]int)
			for _, reading := range tt.readings {
				if reading.ISO3 != "" {
					if currentProgress, exists := countryProgress[reading.ISO3]; !exists || reading.Progresso > currentProgress {
						countryProgress[reading.ISO3] = reading.Progresso
					}
				}
			}

			// Verify aggregated results
			if len(countryProgress) != tt.expectedTotal {
				t.Errorf("Expected %d countries, got %d", tt.expectedTotal, len(countryProgress))
			}

			for iso, expectedProgress := range tt.expectedMax {
				actualProgress, exists := countryProgress[iso]
				if !exists {
					t.Errorf("Expected country %s to exist in results", iso)
					continue
				}
				if actualProgress != expectedProgress {
					t.Errorf("Country %s: expected progress %d, got %d", iso, expectedProgress, actualProgress)
				}
			}
		})
	}
}

func TestStatsResponse_Structure(t *testing.T) {
	countries := []types.CountryProgress{
		{ISO3: "BRA", Progress: 100},
		{ISO3: "USA", Progress: 75},
		{ISO3: "JPN", Progress: 50},
	}

	response := types.StatsResponse{
		Countries: countries,
		Total:     len(countries),
	}

	if response.Total != 3 {
		t.Errorf("Expected total 3, got %d", response.Total)
	}

	if len(response.Countries) != 3 {
		t.Errorf("Expected 3 countries, got %d", len(response.Countries))
	}

	// Test marshaling
	responseBody, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Test unmarshaling
	var unmarshaled types.StatsResponse
	err = json.Unmarshal(responseBody, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.Total != response.Total {
		t.Errorf("Total mismatch after unmarshal: expected %d, got %d", response.Total, unmarshaled.Total)
	}

	if len(unmarshaled.Countries) != len(response.Countries) {
		t.Errorf("Countries length mismatch after unmarshal")
	}
}

func TestStatsResponse_EmptyCountries(t *testing.T) {
	response := types.StatsResponse{
		Countries: []types.CountryProgress{},
		Total:     0,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal empty response: %v", err)
	}

	var unmarshaled types.StatsResponse
	err = json.Unmarshal(responseBody, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty response: %v", err)
	}

	if unmarshaled.Total != 0 {
		t.Errorf("Expected total 0, got %d", unmarshaled.Total)
	}

	if len(unmarshaled.Countries) != 0 {
		t.Errorf("Expected empty countries array, got %d items", len(unmarshaled.Countries))
	}
}

func TestCountryProgress_Structure(t *testing.T) {
	cp := types.CountryProgress{
		ISO3:     "BRA",
		Progress: 75,
	}

	if cp.ISO3 != "BRA" {
		t.Errorf("Expected ISO3 'BRA', got '%s'", cp.ISO3)
	}

	if cp.Progress != 75 {
		t.Errorf("Expected Progress 75, got %d", cp.Progress)
	}

	// Test JSON marshaling
	data, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("Failed to marshal CountryProgress: %v", err)
	}

	var unmarshaled types.CountryProgress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal CountryProgress: %v", err)
	}

	if unmarshaled.ISO3 != cp.ISO3 {
		t.Errorf("ISO3 mismatch after unmarshal")
	}

	if unmarshaled.Progress != cp.Progress {
		t.Errorf("Progress mismatch after unmarshal")
	}
}

func TestProgressEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		progress int
		valid    bool
	}{
		{
			name:     "Zero progress",
			progress: 0,
			valid:    true,
		},
		{
			name:     "100% progress",
			progress: 100,
			valid:    true,
		},
		{
			name:     "Partial progress",
			progress: 50,
			valid:    true,
		},
		{
			name:     "Over 100%",
			progress: 150,
			valid:    true, // System accepts it even if logically invalid
		},
		{
			name:     "Negative progress",
			progress: -10,
			valid:    true, // System accepts it even if logically invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := types.LeituraItem{
				ISO3:      "BRA",
				Progresso: tt.progress,
			}

			// Simulate aggregation
			countryProgress := make(map[string]int)
			if reading.ISO3 != "" {
				countryProgress[reading.ISO3] = reading.Progresso
			}

			if tt.valid && len(countryProgress) == 0 {
				t.Error("Expected progress to be recorded")
			}

			if len(countryProgress) > 0 {
				actualProgress := countryProgress[reading.ISO3]
				if actualProgress != tt.progress {
					t.Errorf("Expected progress %d, got %d", tt.progress, actualProgress)
				}
			}
		})
	}
}

func TestMapToSliceConversion(t *testing.T) {
	countryProgress := map[string]int{
		"BRA": 100,
		"USA": 75,
		"JPN": 50,
		"FRA": 80,
		"DEU": 60,
	}

	countries := make([]types.CountryProgress, 0, len(countryProgress))
	for iso, progress := range countryProgress {
		countries = append(countries, types.CountryProgress{
			ISO3:     iso,
			Progress: progress,
		})
	}

	if len(countries) != len(countryProgress) {
		t.Errorf("Expected %d countries in slice, got %d", len(countryProgress), len(countries))
	}

	// Verify all countries from map are in slice
	foundCountries := make(map[string]bool)
	for _, country := range countries {
		foundCountries[country.ISO3] = true
		expectedProgress := countryProgress[country.ISO3]
		if country.Progress != expectedProgress {
			t.Errorf("Country %s: expected progress %d, got %d", country.ISO3, expectedProgress, country.Progress)
		}
	}

	for iso := range countryProgress {
		if !foundCountries[iso] {
			t.Errorf("Country %s missing from slice", iso)
		}
	}
}

func TestResponseJSONFormat(t *testing.T) {
	response := types.StatsResponse{
		Countries: []types.CountryProgress{
			{ISO3: "BRA", Progress: 100},
			{ISO3: "USA", Progress: 75},
		},
		Total: 2,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := jsonMap["countries"]; !exists {
		t.Error("Expected 'countries' field in JSON")
	}

	if _, exists := jsonMap["total"]; !exists {
		t.Error("Expected 'total' field in JSON")
	}

	// Verify countries is an array
	countries, ok := jsonMap["countries"].([]interface{})
	if !ok {
		t.Error("Expected 'countries' to be an array")
	}

	if len(countries) != 2 {
		t.Errorf("Expected 2 countries in JSON, got %d", len(countries))
	}

	// Verify total is a number
	total, ok := jsonMap["total"].(float64)
	if !ok {
		t.Error("Expected 'total' to be a number")
	}

	if int(total) != 2 {
		t.Errorf("Expected total 2, got %d", int(total))
	}
}

func BenchmarkCountryProgressAggregation(b *testing.B) {
	// Create test data
	readings := make([]types.LeituraItem, 1000)
	countries := []string{"BRA", "USA", "JPN", "FRA", "DEU", "GBR", "ITA", "ESP", "CAN", "AUS"}

	for i := 0; i < 1000; i++ {
		readings[i] = types.LeituraItem{
			ISO3:      countries[i%len(countries)],
			Progresso: i % 100,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countryProgress := make(map[string]int)
		for _, reading := range readings {
			if reading.ISO3 != "" {
				if currentProgress, exists := countryProgress[reading.ISO3]; !exists || reading.Progresso > currentProgress {
					countryProgress[reading.ISO3] = reading.Progresso
				}
			}
		}
	}
}

func BenchmarkStatsResponseMarshal(b *testing.B) {
	countries := make([]types.CountryProgress, 100)
	for i := 0; i < 100; i++ {
		countries[i] = types.CountryProgress{
			ISO3:     "XXX",
			Progress: i,
		}
	}

	response := types.StatsResponse{
		Countries: countries,
		Total:     len(countries),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(response)
		if err != nil {
			b.Fatal(err)
		}
	}
}
