package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mundotalendo/functions/mapping"
	"github.com/mundotalendo/functions/types"
	"github.com/mundotalendo/functions/utils"
)

// ValidDesafioTypes defines the types of desafios we process.
var ValidDesafioTypes = map[string]bool{
	"leitura":   true,
	"atividade": true,
}

// ProcessingMeta contains metadata for processing a webhook.
type ProcessingMeta struct {
	UUID      string    // Webhook UUID
	User      string    // User name
	AvatarURL string    // User avatar URL
	Timestamp time.Time // Processing timestamp
}

// ProcessingResult contains the result of processing a desafio.
type ProcessingResult struct {
	ISO3      string // Country ISO3 code
	Country   string // Country name
	Processed bool   // Whether processing succeeded
	Error     error  // Error if processing failed
}

// DesafioProcessor handles the processing of reading challenges.
type DesafioProcessor struct {
	store *LeituraStore
}

// NewDesafioProcessor creates a new processor with the given DynamoDB store.
func NewDesafioProcessor(store *LeituraStore) *DesafioProcessor {
	return &DesafioProcessor{
		store: store,
	}
}

// ProcessAll processes all desafios in a payload and saves them to DynamoDB.
// It returns the count of successfully processed items and errors.
//
// Note: This function continues processing even if individual desafios fail.
// Only transient errors (DynamoDB throttling) should cause a retry.
func (p *DesafioProcessor) ProcessAll(ctx context.Context, payload *types.WebhookPayload, meta ProcessingMeta) (int, int, []ProcessingResult) {
	processed := 0
	errors := 0
	results := make([]ProcessingResult, 0, len(payload.Desafios))

	for i, desafio := range payload.Desafios {
		result := p.processDesafio(ctx, desafio, i, payload, meta)
		results = append(results, result)

		if result.Processed {
			processed++
		} else if result.Error != nil {
			errors++
		}
	}

	return processed, errors, results
}

// processDesafio handles a single desafio, extracting data and saving to DynamoDB.
func (p *DesafioProcessor) processDesafio(ctx context.Context, desafio types.Desafio, index int, payload *types.WebhookPayload, meta ProcessingMeta) ProcessingResult {
	// Filter: only process valid types
	if !ValidDesafioTypes[desafio.Tipo] {
		return ProcessingResult{Processed: false}
	}

	// Extract progress and book data
	progress, latestUpdate, bookTitle, capaURL := extractDesafioData(desafio)

	// Clean emojis from country name and category
	cleanedCountry := utils.CleanEmojis(desafio.Descricao)
	cleanedCategory := utils.CleanEmojis(desafio.Categoria)

	// Map country to ISO3
	iso3 := mapping.GetISO(cleanedCountry)
	if iso3 == "" {
		log.Printf("Country not found: %s (original: %s)", cleanedCountry, desafio.Descricao)
		return ProcessingResult{
			Country: cleanedCountry,
			Error:   fmt.Errorf("%w: %s", ErrCountryNotFound, cleanedCountry),
		}
	}

	// Create LeituraItem
	item := types.LeituraItem{
		PK:          "EVENT#LEITURA",
		SK:          fmt.Sprintf("%s#%s#%d", meta.UUID, iso3, index),
		ISO3:        iso3,
		Pais:        cleanedCountry,
		Categoria:   cleanedCategory,
		Progresso:   progress,
		User:        meta.User,
		ImagemURL:   meta.AvatarURL,
		CapaURL:     capaURL,
		Livro:       bookTitle,
		WebhookUUID: meta.UUID,
		UpdatedAt:   latestUpdate.Format(time.RFC3339),
	}

	// Save to DynamoDB
	if err := p.store.SaveLeitura(ctx, item); err != nil {
		log.Printf("ERROR saving leitura for %s: %v", iso3, err)
		return ProcessingResult{
			ISO3:    iso3,
			Country: cleanedCountry,
			Error:   err,
		}
	}

	log.Printf("Processed: %s (%s) - User: %s - Category: %s - Progress: %d%%",
		cleanedCountry, iso3, meta.User, cleanedCategory, progress)

	return ProcessingResult{
		ISO3:      iso3,
		Country:   cleanedCountry,
		Processed: true,
	}
}

// extractDesafioData extracts progress, update timestamp, book title, and cover URL from a desafio.
// It handles multiple vinculados by taking the maximum progress and latest update.
func extractDesafioData(desafio types.Desafio) (progress int, latestUpdate time.Time, bookTitle, capaURL string) {
	maxProgress := 0
	latestUpdate = time.Time{}

	for _, vinculado := range desafio.Vinculados {
		// Track maximum progress
		if vinculado.Progresso > maxProgress {
			maxProgress = vinculado.Progresso
		}

		// Extract book title
		if vinculado.Edicao != nil && vinculado.Edicao.Titulo != "" {
			bookTitle = vinculado.Edicao.Titulo
		}

		// Extract cover URL
		if vinculado.Edicao != nil && vinculado.Edicao.Capa != "" {
			capaURL = vinculado.Edicao.Capa
		}

		// Parse update timestamp
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

	// Handle empty vinculados
	if len(desafio.Vinculados) == 0 {
		latestUpdate = time.Now()
	}

	// If concluido == true, force progress to 100%
	if desafio.Concluido {
		maxProgress = 100
	}

	// Clamp progress to 0-100
	progress = clampProgress(maxProgress)

	return progress, latestUpdate, bookTitle, capaURL
}

// clampProgress ensures progress is within valid range [0, 100].
func clampProgress(progress int) int {
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}
	return progress
}
