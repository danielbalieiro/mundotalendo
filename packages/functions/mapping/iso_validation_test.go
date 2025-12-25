package mapping

import (
	"fmt"
	"testing"
)

// TestValidateAllMaratonaCountries valida todos os 195 países da lista oficial do Maratona.app
func TestValidateAllMaratonaCountries(t *testing.T) {
	// Lista completa dos 195 países do Maratona.app 2026
	// Organizada por mês para facilitar identificação
	marathonaCountries := map[string][]string{
		"Janeiro": {
			"Brasil", "Guiana Francesa", "Suriname", "Guiana", "Venezuela",
			"Colômbia", "Equador", "Peru", "Bolívia", "Chile",
			"Paraguai", "Argentina", "Uruguai",
		},
		"Fevereiro": {
			"China", "Japão", "Coreia do Sul", "Coreia do Norte", "Filipinas",
			"Indonésia", "Butão", "Mongólia", "Laos", "Nepal",
			"Vietnã", "Brunei", "Malásia", "Timor Leste", "Cazaquistão",
			"Camboja", "Tailândia", "Mianmar", "Singapura", "Taiwan",
		},
		"Março": {
			"Portugal", "Espanha", "França", "Andorra", "Mônaco",
			"Itália", "Malta", "Vaticano", "San Marino",
		},
		"Abril": {
			"Guiné Equatorial", "Gabão", "Congo", "República Democrática do Congo", "Uganda",
			"Quênia", "Ruanda", "Burundi", "Tanzânia", "Angola",
			"Zâmbia", "Malawi", "Moçambique", "Zimbábue", "Botsuana",
			"Namíbia", "África do Sul", "Lesoto", "Essuatíni", "Madagascar",
			"São Tomé e Príncipe", "Seychelles", "Comores",
		},
		"Maio": {
			"Guatemala", "Belize", "El Salvador", "Honduras", "Nicarágua",
			"Costa Rica", "Panamá", "Bahamas", "Cuba", "Jamaica",
			"Haiti", "República Dominicana", "Porto Rico", "São Cristóvão e Névis", "Antígua e Barbuda",
			"Montserrat", "Dominica", "Santa Lúcia", "Barbados", "Granada",
			"Trindade e Tobago", "São Vicente e Grandinas",
		},
		"Junho": {
			"Inglaterra", "Irlanda", "Islândia", "Noruega", "Suécia",
			"Finlândia", "Escócia", "País de Gales", "Irlanda do norte",
		},
		"Julho": {
			"Canadá", "Estados Unidos", "México", "Groelândia",
		},
		"Agosto": {
			"Austrália", "Papua-Nova Guiné", "Nova Zelândia", "Fiji", "Ilhas Salomão",
			"Vanuatu", "Samoa", "Kiribati", "Tonga", "Micronésia",
			"Palau", "Ilhas Marshall", "Nauru", "Tuvalu",
		},
		"Setembro": {
			"Suiça", "Bélgica", "Luxemburgo", "Países Baixos", "Alemanha",
			"Dinamarca", "Polônia", "Tchéquia", "Áustria", "Liechtenstein",
		},
		"Outubro": {
			"Eslováquia", "Hungria", "Eslovênia", "Croácia", "Bósnia-Herzegóvina",
			"Montenegro", "Sérvia", "Albânia", "Grécia", "Macedônia do Norte",
			"Bulgária", "Romênia", "Moldávia", "Ucrânia", "Bielorrússia",
			"Lituânia", "Letônia", "Estônia", "Rússia",
		},
		"Novembro": {
			"Marrocos", "Argélia", "Tunísia", "Saara Ocidental", "Mauritânia",
			"Senegal", "Gâmbia", "Guiné-Bissau", "Guiné", "Serra Leoa",
			"Libéria", "Costa do Marfim", "Mali", "Burkina Faso", "Gana",
			"Togo", "Benin", "Níger", "Nigéria", "Líbia",
			"Chade", "Camarões", "República Centro-Africana", "Egito", "Sudão",
			"Sudão do Sul", "Etiópia", "Somália", "Eritreia", "Djibouti",
			"Cabo verde",
		},
		"Dezembro": {
			"Turquia", "Chipre", "Líbano", "Israel", "Palestina",
			"Jordânia", "Síria", "Iraque", "Irã", "Geórgia",
			"Armênia", "Azerbajão", "Turcomenistão", "Uzbequistão", "Afeganistão",
			"Tajiquistão", "Quirguistão", "Paquistão", "Arábia Saudita", "Kuwait",
			"Bahrein", "Catar", "Emirados Árabes", "Omã", "Iêmen",
			"Índia", "Sri Lanka", "Maldivas", "Bangladesh",
		},
	}

	var allCountries []string
	var unmappedCountries []string
	var mappedCountries []string

	// Processar cada mês
	for month, countries := range marathonaCountries {
		for _, country := range countries {
			allCountries = append(allCountries, country)
			iso := GetISO(country)
			if iso == "" {
				unmappedCountries = append(unmappedCountries, fmt.Sprintf("%s (%s)", country, month))
			} else {
				mappedCountries = append(mappedCountries, fmt.Sprintf("%s → %s", country, iso))
			}
		}
	}

	// Gerar relatório
	totalCountries := len(allCountries)
	mapped := len(mappedCountries)
	unmapped := len(unmappedCountries)
	mappedPercentage := float64(mapped) / float64(totalCountries) * 100

	t.Logf("\n")
	t.Logf("╔═══════════════════════════════════════════════════════════════════════╗")
	t.Logf("║       VALIDAÇÃO DE PAÍSES - MARATONA.APP 2026                        ║")
	t.Logf("╠═══════════════════════════════════════════════════════════════════════╣")
	t.Logf("║ Total de países:        %3d                                           ║", totalCountries)
	t.Logf("║ Mapeados:               %3d (%.1f%%)                                  ║", mapped, mappedPercentage)
	t.Logf("║ Não mapeados:           %3d (%.1f%%)                                  ║", unmapped, 100-mappedPercentage)
	t.Logf("╚═══════════════════════════════════════════════════════════════════════╝")
	t.Logf("\n")

	if unmapped > 0 {
		t.Logf("⚠️  PAÍSES NÃO MAPEADOS:")
		t.Logf("═══════════════════════════════════════════════")
		for _, country := range unmappedCountries {
			t.Logf("  ❌ %s", country)
		}
		t.Logf("\n")
		t.Logf("💡 RECOMENDAÇÃO: Adicionar aliases para esses países em iso.go")
		t.Logf("\n")
	}

	if unmapped > 0 {
		t.Errorf("Validação FALHOU: %d países não estão mapeados. Execute este teste para ver a lista completa.", unmapped)
	} else {
		t.Logf("✅ VALIDAÇÃO COMPLETA: Todos os %d países estão corretamente mapeados!", totalCountries)
	}
}

// TestCountryVariations testa variações conhecidas de grafia
func TestCountryVariations(t *testing.T) {
	variations := map[string]string{
		// Variações de acento
		"Suiça":          "CHE", // vs Suíça
		"Azerbajão":      "AZE", // vs Azerbaijão
		// Variações de grafia
		"Groelândia":     "GRL", // vs Groenlândia
		// Variações de capitalização
		"irlanda do norte": "GBR", // vs Irlanda do Norte (se case-insensitive)
	}

	for country, expectedISO := range variations {
		iso := GetISO(country)
		if iso == "" {
			t.Errorf("Variação não mapeada: '%s' (esperado: %s)", country, expectedISO)
		} else if iso != expectedISO {
			t.Errorf("Variação mapeada incorretamente: '%s' → %s (esperado: %s)", country, iso, expectedISO)
		} else {
			t.Logf("✅ Variação correta: '%s' → %s", country, iso)
		}
	}
}
