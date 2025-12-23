package main

import (
	"testing"

	"github.com/mundotalendo/functions/types"
)

func TestFindLatestPerUser(t *testing.T) {
	items := []types.LeituraItem{
		{User: "Alice", SK: "TIMESTAMP#2025-12-20T10:00:00Z#0", ISO3: "BRA", Pais: "Brasil", ImagemURL: "https://example.com/alice1.jpg"},
		{User: "Alice", SK: "TIMESTAMP#2025-12-23T15:00:00Z#0", ISO3: "PRT", Pais: "Portugal", ImagemURL: "https://example.com/alice2.jpg"},
		{User: "Bob", SK: "TIMESTAMP#2025-12-22T12:00:00Z#0", ISO3: "FRA", Pais: "França", ImagemURL: "https://example.com/bob.jpg"},
		{User: "Charlie", SK: "TIMESTAMP#2025-12-21T08:00:00Z#0", ISO3: "DEU", Pais: "Alemanha", ImagemURL: "https://example.com/charlie.jpg"},
	}

	// Simulate finding latest per user
	userLatest := make(map[string]types.LeituraItem)
	for _, item := range items {
		if existing, exists := userLatest[item.User]; !exists || item.SK > existing.SK {
			userLatest[item.User] = item
		}
	}

	// Test: Alice should have Portugal (most recent)
	if aliceLatest, exists := userLatest["Alice"]; exists {
		if aliceLatest.ISO3 != "PRT" {
			t.Errorf("Expected Alice's latest country to be PRT, got %s", aliceLatest.ISO3)
		}
		if aliceLatest.SK != "TIMESTAMP#2025-12-23T15:00:00Z#0" {
			t.Errorf("Expected Alice's latest SK to be 2025-12-23, got %s", aliceLatest.SK)
		}
	} else {
		t.Error("Expected to find Alice in userLatest map")
	}

	// Test: Bob should have France
	if bobLatest, exists := userLatest["Bob"]; exists {
		if bobLatest.ISO3 != "FRA" {
			t.Errorf("Expected Bob's latest country to be FRA, got %s", bobLatest.ISO3)
		}
	} else {
		t.Error("Expected to find Bob in userLatest map")
	}

	// Test: Should have exactly 3 users
	if len(userLatest) != 3 {
		t.Errorf("Expected 3 unique users, got %d", len(userLatest))
	}
}

func TestEmptyAvatarURL(t *testing.T) {
	items := []types.LeituraItem{
		{User: "Alice", SK: "TIMESTAMP#2025-12-23T10:00:00Z#0", ISO3: "BRA", ImagemURL: "https://example.com/alice.jpg"},
		{User: "Bob", SK: "TIMESTAMP#2025-12-23T11:00:00Z#0", ISO3: "PRT", ImagemURL: ""}, // Empty avatar
		{User: "Charlie", SK: "TIMESTAMP#2025-12-23T12:00:00Z#0", ISO3: "FRA", ImagemURL: "https://example.com/charlie.jpg"},
	}

	// Build user locations (simulating what the handler does)
	userLatest := make(map[string]types.LeituraItem)
	for _, item := range items {
		if item.User == "" {
			continue // Skip empty users
		}
		if existing, exists := userLatest[item.User]; !exists || item.SK > existing.SK {
			userLatest[item.User] = item
		}
	}

	users := make([]types.UserLocation, 0, len(userLatest))
	for userName, item := range userLatest {
		users = append(users, types.UserLocation{
			User:      userName,
			AvatarURL: item.ImagemURL,
			ISO3:      item.ISO3,
			Pais:      item.Pais,
			Timestamp: item.SK,
		})
	}

	// Test: Should have 3 users (including Bob with empty avatar)
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Test: Bob should have empty avatarURL
	bobFound := false
	for _, user := range users {
		if user.User == "Bob" {
			bobFound = true
			if user.AvatarURL != "" {
				t.Errorf("Expected Bob to have empty avatarURL, got %s", user.AvatarURL)
			}
		}
	}
	if !bobFound {
		t.Error("Expected to find Bob in users list")
	}
}

func TestSKComparison(t *testing.T) {
	// Test that SK lexicographic comparison works correctly for timestamps
	tests := []struct {
		name     string
		sk1      string
		sk2      string
		expected string // Which SK should be "greater" (more recent)
	}{
		{
			name:     "Different dates",
			sk1:      "TIMESTAMP#2025-12-20T10:00:00Z#0",
			sk2:      "TIMESTAMP#2025-12-23T10:00:00Z#0",
			expected: "sk2",
		},
		{
			name:     "Same date, different times",
			sk1:      "TIMESTAMP#2025-12-23T10:00:00Z#0",
			sk2:      "TIMESTAMP#2025-12-23T15:00:00Z#0",
			expected: "sk2",
		},
		{
			name:     "Same timestamp, different index",
			sk1:      "TIMESTAMP#2025-12-23T15:00:00Z#0",
			sk2:      "TIMESTAMP#2025-12-23T15:00:00Z#1",
			expected: "sk2",
		},
		{
			name:     "Identical SKs",
			sk1:      "TIMESTAMP#2025-12-23T15:00:00Z#0",
			sk2:      "TIMESTAMP#2025-12-23T15:00:00Z#0",
			expected: "equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected == "sk2" && !(tt.sk2 > tt.sk1) {
				t.Errorf("Expected sk2 (%s) to be greater than sk1 (%s)", tt.sk2, tt.sk1)
			}
			if tt.expected == "sk1" && !(tt.sk1 > tt.sk2) {
				t.Errorf("Expected sk1 (%s) to be greater than sk2 (%s)", tt.sk1, tt.sk2)
			}
			if tt.expected == "equal" && tt.sk1 != tt.sk2 {
				t.Errorf("Expected sk1 and sk2 to be equal, got %s != %s", tt.sk1, tt.sk2)
			}
		})
	}
}

func TestUserLocationStructure(t *testing.T) {
	userLocation := types.UserLocation{
		User:      "TestUser",
		AvatarURL: "https://example.com/avatar.jpg",
		ISO3:      "BRA",
		Pais:      "Brasil",
		Timestamp: "TIMESTAMP#2025-12-23T15:00:00Z#0",
	}

	// Test all fields are set correctly
	if userLocation.User != "TestUser" {
		t.Errorf("Expected User 'TestUser', got '%s'", userLocation.User)
	}
	if userLocation.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("Expected AvatarURL to be set, got '%s'", userLocation.AvatarURL)
	}
	if userLocation.ISO3 != "BRA" {
		t.Errorf("Expected ISO3 'BRA', got '%s'", userLocation.ISO3)
	}
	if userLocation.Pais != "Brasil" {
		t.Errorf("Expected Pais 'Brasil', got '%s'", userLocation.Pais)
	}
	if userLocation.Timestamp == "" {
		t.Error("Expected Timestamp to be set")
	}
}

func TestUserLocationsResponse(t *testing.T) {
	users := []types.UserLocation{
		{User: "Alice", AvatarURL: "https://example.com/alice.jpg", ISO3: "BRA", Pais: "Brasil"},
		{User: "Bob", AvatarURL: "https://example.com/bob.jpg", ISO3: "PRT", Pais: "Portugal"},
	}

	response := types.UserLocationsResponse{
		Users: users,
		Total: len(users),
	}

	if response.Total != 2 {
		t.Errorf("Expected Total to be 2, got %d", response.Total)
	}
	if len(response.Users) != 2 {
		t.Errorf("Expected 2 users in response, got %d", len(response.Users))
	}
}

func TestEmptyUserName(t *testing.T) {
	items := []types.LeituraItem{
		{User: "Alice", SK: "TIMESTAMP#2025-12-23T10:00:00Z#0", ISO3: "BRA"},
		{User: "", SK: "TIMESTAMP#2025-12-23T11:00:00Z#0", ISO3: "PRT"}, // Empty user name
		{User: "Charlie", SK: "TIMESTAMP#2025-12-23T12:00:00Z#0", ISO3: "FRA"},
	}

	// Simulate filtering out empty user names
	userLatest := make(map[string]types.LeituraItem)
	for _, item := range items {
		if item.User == "" {
			continue // Skip empty users (as the handler does)
		}
		if existing, exists := userLatest[item.User]; !exists || item.SK > existing.SK {
			userLatest[item.User] = item
		}
	}

	// Test: Should have only 2 users (empty user filtered out)
	if len(userLatest) != 2 {
		t.Errorf("Expected 2 users after filtering, got %d", len(userLatest))
	}

	// Test: Empty user should not be in map
	if _, exists := userLatest[""]; exists {
		t.Error("Expected empty user name to be filtered out")
	}
}

func BenchmarkFindLatestPerUser(b *testing.B) {
	// Create a larger dataset for benchmarking
	items := make([]types.LeituraItem, 0, 1000)
	for i := 0; i < 1000; i++ {
		items = append(items, types.LeituraItem{
			User: "User" + string(rune(i%100)),
			SK:   "TIMESTAMP#2025-12-23T10:00:00Z#" + string(rune(i)),
			ISO3: "BRA",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userLatest := make(map[string]types.LeituraItem)
		for _, item := range items {
			if item.User == "" {
				continue
			}
			if existing, exists := userLatest[item.User]; !exists || item.SK > existing.SK {
				userLatest[item.User] = item
			}
		}
	}
}
