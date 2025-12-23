package mapping

import (
	"testing"
)

func TestGetISO(t *testing.T) {
	tests := []struct {
		name         string
		countryName  string
		expectedISO  string
		description  string
	}{
		{
			name:         "Brasil",
			countryName:  "Brasil",
			expectedISO:  "BRA",
			description:  "Should return BRA for Brasil",
		},
		{
			name:         "Estados Unidos",
			countryName:  "Estados Unidos",
			expectedISO:  "USA",
			description:  "Should return USA for Estados Unidos",
		},
		{
			name:         "Alasca",
			countryName:  "Alasca",
			expectedISO:  "USA",
			description:  "Should return USA for Alasca",
		},
		{
			name:         "Portugal",
			countryName:  "Portugal",
			expectedISO:  "PRT",
			description:  "Should return PRT for Portugal",
		},
		{
			name:         "França",
			countryName:  "França",
			expectedISO:  "FRA",
			description:  "Should return FRA for França",
		},
		{
			name:         "Japão",
			countryName:  "Japão",
			expectedISO:  "JPN",
			description:  "Should return JPN for Japão",
		},
		{
			name:         "China",
			countryName:  "China",
			expectedISO:  "CHN",
			description:  "Should return CHN for China",
		},
		{
			name:         "Canadá",
			countryName:  "Canadá",
			expectedISO:  "CAN",
			description:  "Should return CAN for Canadá",
		},
		{
			name:         "Unknown country",
			countryName:  "País Inexistente",
			expectedISO:  "",
			description:  "Should return empty string for unknown country",
		},
		{
			name:         "Empty string",
			countryName:  "",
			expectedISO:  "",
			description:  "Should return empty string for empty input",
		},
		{
			name:         "País with special characters",
			countryName:  "São Tomé e Príncipe",
			expectedISO:  "STP",
			description:  "Should handle special characters",
		},
		{
			name:         "País with accents",
			countryName:  "Países Baixos",
			expectedISO:  "NLD",
			description:  "Should handle accents correctly",
		},
		{
			name:         "República Democrática do Congo",
			countryName:  "República Democrática do Congo",
			expectedISO:  "COD",
			description:  "Should handle long names",
		},
		{
			name:         "Reino Unido",
			countryName:  "Reino Unido",
			expectedISO:  "GBR",
			description:  "Should return GBR for Reino Unido",
		},
		{
			name:         "Inglaterra",
			countryName:  "Inglaterra",
			expectedISO:  "GBR",
			description:  "Should return GBR for Inglaterra (UK alias)",
		},
		{
			name:         "Escócia",
			countryName:  "Escócia",
			expectedISO:  "GBR",
			description:  "Should return GBR for Escócia (UK alias)",
		},
		{
			name:         "País de Gales",
			countryName:  "País de Gales",
			expectedISO:  "GBR",
			description:  "Should return GBR for País de Gales (UK alias)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetISO(tt.countryName)
			if result != tt.expectedISO {
				t.Errorf("%s: got %s, want %s", tt.description, result, tt.expectedISO)
			}
		})
	}
}

func TestGetISO_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name        string
		countryName string
		shouldFind  bool
	}{
		{
			name:        "Exact match",
			countryName: "Brasil",
			shouldFind:  true,
		},
		{
			name:        "All lowercase",
			countryName: "brasil",
			shouldFind:  false,
		},
		{
			name:        "All uppercase",
			countryName: "BRASIL",
			shouldFind:  false,
		},
		{
			name:        "Mixed case",
			countryName: "BrAsIl",
			shouldFind:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetISO(tt.countryName)
			if tt.shouldFind && result == "" {
				t.Errorf("Expected to find ISO for %s, but got empty string", tt.countryName)
			}
			if !tt.shouldFind && result != "" {
				t.Errorf("Expected empty string for %s, but got %s", tt.countryName, result)
			}
		})
	}
}

func TestNameToIso_DataIntegrity(t *testing.T) {
	t.Run("All ISO codes are 3 letters", func(t *testing.T) {
		for countryName, iso := range NameToIso {
			if len(iso) != 3 {
				t.Errorf("Country %s has ISO code %s with length %d, expected 3", countryName, iso, len(iso))
			}
		}
	})

	t.Run("All ISO codes are uppercase", func(t *testing.T) {
		for countryName, iso := range NameToIso {
			for _, char := range iso {
				if char < 'A' || char > 'Z' {
					t.Errorf("Country %s has ISO code %s with non-uppercase character", countryName, iso)
				}
			}
		}
	})

	t.Run("No empty country names", func(t *testing.T) {
		for countryName := range NameToIso {
			if countryName == "" {
				t.Error("Found empty country name in NameToIso map")
			}
		}
	})

	t.Run("No empty ISO codes", func(t *testing.T) {
		for countryName, iso := range NameToIso {
			if iso == "" {
				t.Errorf("Country %s has empty ISO code", countryName)
			}
		}
	})

	t.Run("Map has substantial number of countries", func(t *testing.T) {
		count := len(NameToIso)
		if count < 150 {
			t.Errorf("Expected at least 150 countries, got %d", count)
		}
	})
}

func TestNameToIso_SpecificMappings(t *testing.T) {
	expectedMappings := map[string]string{
		"Brasil":                        "BRA",
		"Estados Unidos":                "USA",
		"Alasca":                        "USA",
		"Portugal":                      "PRT",
		"França":                        "FRA",
		"Alemanha":                      "DEU",
		"Japão":                         "JPN",
		"China":                         "CHN",
		"Canadá":                        "CAN",
		"Reino Unido":                   "GBR",
		"Rússia":                        "RUS",
		"Índia":                         "IND",
		"África do Sul":                 "ZAF",
		"Austrália":                     "AUS",
		"Argentina":                     "ARG",
		"México":                        "MEX",
		"Coreia do Sul":                 "KOR",
		"Coreia do Norte":               "PRK",
		"Itália":                        "ITA",
		"Espanha":                       "ESP",
		"Emirados Árabes Unidos":        "ARE",
		"Arábia Saudita":                "SAU",
		"República Democrática do Congo": "COD",
		"Congo":                         "COG",
	}

	for countryName, expectedISO := range expectedMappings {
		t.Run(countryName, func(t *testing.T) {
			actualISO := NameToIso[countryName]
			if actualISO != expectedISO {
				t.Errorf("For %s: expected %s, got %s", countryName, expectedISO, actualISO)
			}
		})
	}
}

func TestNameToIso_NoISOCollisions(t *testing.T) {
	// Check for potential issues where different countries map to same ISO
	// (except for special cases like Alaska -> USA)
	isoToCountries := make(map[string][]string)

	for countryName, iso := range NameToIso {
		isoToCountries[iso] = append(isoToCountries[iso], countryName)
	}

	knownDuplicates := map[string]bool{
		"USA": true, // Alasca and Estados Unidos
		"GBR": true, // Reino Unido, Inglaterra, Escócia, País de Gales, Irlanda do Norte
		"SWZ": true, // Essuatíni and Suazilândia (same country, different names)
	}

	for iso, countries := range isoToCountries {
		if len(countries) > 1 && !knownDuplicates[iso] {
			t.Logf("Multiple countries map to ISO %s: %v", iso, countries)
		}
	}
}

func BenchmarkGetISO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetISO("Brasil")
	}
}

func BenchmarkGetISO_NotFound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetISO("País Inexistente")
	}
}

func BenchmarkGetISO_LongName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetISO("República Democrática do Congo")
	}
}
