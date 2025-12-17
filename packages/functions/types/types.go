package types

// Webhook payload structures
type Maratona struct {
	Nome          string `json:"nome"`
	Identificador string `json:"identificador"`
}

type WebhookPayload struct {
	Perfil   Perfil     `json:"perfil"`
	Maratona Maratona   `json:"maratona"`
	Desafios []Desafio `json:"desafios"`
}

type Perfil struct {
	Nome string `json:"nome"`
	Link string `json:"link"`
}

type Desafio struct {
	ID         string      `json:"id,omitempty"`
	Descricao  string      `json:"descricao"`
	Categoria  string      `json:"categoria"`
	Concluido  bool        `json:"concluido"`
	Tipo       string      `json:"tipo"`
	Vinculados []Vinculado `json:"vinculados"`
}

type Edicao struct {
	Titulo string `json:"titulo,omitempty"`
	Autor  string `json:"autor,omitempty"`
	Capa   string `json:"capa,omitempty"`
}

type Vinculado struct {
	ID         string  `json:"id,omitempty"`
	Completo   bool    `json:"completo"`
	Progresso  int     `json:"progresso"`
	Avaliacao  int     `json:"avaliacao,omitempty"`
	Comentario string  `json:"comentario,omitempty"`
	UpdatedAt  string  `json:"updatedAt"`
	Edicao     *Edicao `json:"edicao,omitempty"`
	DiaMarcado string  `json:"diaMarcado,omitempty"`
}

// DynamoDB item structure
type LeituraItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	ISO3      string `dynamodbav:"iso3"`
	Pais      string `dynamodbav:"pais"`
	Categoria string `dynamodbav:"categoria"`
	Progresso int    `dynamodbav:"progresso"`
	User      string `dynamodbav:"user"`
	Metadata  string `dynamodbav:"metadata"`
}

// Falhas table item structure
type FalhaItem struct {
	PK              string `dynamodbav:"PK"`
	SK              string `dynamodbav:"SK"`
	ErrorType       string `dynamodbav:"errorType"`
	ErrorMessage    string `dynamodbav:"errorMessage"`
	OriginalPayload string `dynamodbav:"originalPayload"`
}

// Stats response structure
type CountryProgress struct {
	ISO3     string `json:"iso3"`
	Progress int    `json:"progress"`
}

type StatsResponse struct {
	Countries []CountryProgress `json:"countries"`
	Total     int               `json:"total"`
}
