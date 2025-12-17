package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mundotalendo/functions/mapping"
	"github.com/mundotalendo/functions/types"
)

var (
	dynamoClient *dynamodb.Client
	tableName    string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	tableName = os.Getenv("SST_Resource_Leituras_name")
	rand.Seed(time.Now().UnixNano())
}

// Get all country names from the mapping
func getAllCountries() []string {
	countries := make([]string, 0, len(mapping.NameToIso))
	for country := range mapping.NameToIso {
		countries = append(countries, country)
	}
	return countries
}

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.Println("Seeding database with random countries")

	// Parse request for count (default 10)
	count := 10
	if request.Body != "" {
		var req struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal([]byte(request.Body), &req); err == nil && req.Count > 0 {
			count = req.Count
			if count > 100 {
				count = 100 // Safety limit
			}
		}
	}

	allCountries := getAllCountries()
	inserted := 0
	categories := []string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}

	for i := 0; i < count; i++ {
		// Random country
		randomCountry := allCountries[rand.Intn(len(allCountries))]
		iso3 := mapping.GetISO(randomCountry)
		if iso3 == "" {
			continue
		}

		// Random category (month)
		randomCategory := categories[rand.Intn(len(categories))]

		// Random timestamp within the last 30 days
		daysAgo := rand.Intn(30)
		timestamp := time.Now().AddDate(0, 0, -daysAgo)

		// Random progress 0-100
		randomProgress := rand.Intn(101)

		// Generate random user
		userName := fmt.Sprintf("TestUser%d", rand.Intn(100))

		// Create sample metadata
		samplePayload := types.WebhookPayload{
			Perfil: types.Perfil{
				Nome: userName,
				Link: fmt.Sprintf("https://maratona.app/user/%s", userName),
			},
			Maratona: types.Maratona{
				Nome:          "Lendo Países",
				Identificador: "maratona-lendo-paises",
			},
			Desafios: []types.Desafio{{
				Descricao: randomCountry,
				Categoria: randomCategory,
				Tipo:      "leitura",
				Vinculados: []types.Vinculado{{
					Progresso: randomProgress,
					UpdatedAt: timestamp,
				}},
			}},
		}

		metadataBytes, _ := json.Marshal(samplePayload)

		item := types.LeituraItem{
			PK:        "EVENT#LEITURA",
			SK:        fmt.Sprintf("TIMESTAMP#%s", timestamp.Format(time.RFC3339)),
			ISO3:      iso3,
			Pais:      randomCountry,
			Categoria: randomCategory,
			Progresso: randomProgress,
			User:      userName,
			Metadata:  string(metadataBytes),
		}

		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Printf("Error marshaling item: %v", err)
			continue
		}

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      av,
		})
		if err != nil {
			log.Printf("Error inserting item: %v", err)
			continue
		}

		log.Printf("Inserted: %s (%s) - Category: %s", randomCountry, iso3, randomCategory)
		inserted++
	}

	response := map[string]interface{}{
		"success":  true,
		"inserted": inserted,
		"message":  fmt.Sprintf("Inserted %d random readings", inserted),
	}

	responseBody, _ := json.Marshal(response)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
