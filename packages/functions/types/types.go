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
	Nome   string `json:"nome"`
	Link   string `json:"link"`
	Imagem string `json:"imagem"` // URL do avatar do usuário
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

// DynamoDB item structures

// LeituraItem - Item de leitura (país) agrupado por UUID do webhook
// PK: "EVENT#LEITURA#<uuid>" - agrupa todos os eventos do mesmo webhook
// SK: "COUNTRY#<iso3>" - identifica o país dentro do webhook
type LeituraItem struct {
	PK        string `dynamodbav:"PK"`        // "EVENT#LEITURA#<uuid>"
	SK        string `dynamodbav:"SK"`        // "COUNTRY#<iso3>"
	ISO3      string `dynamodbav:"iso3"`      // Código ISO 3166-1 Alpha-3
	Pais      string `dynamodbav:"pais"`      // Nome do país em português
	Categoria string `dynamodbav:"categoria"` // Mês/categoria do desafio
	Progresso int    `dynamodbav:"progresso"` // Progresso 0-100%
	User      string `dynamodbav:"user"`      // Nome do usuário
	ImagemURL string `dynamodbav:"imagemURL"` // URL do avatar do usuário
	Livro     string `dynamodbav:"livro"`     // Título do livro sendo lido
	// Metadata REMOVIDO - payload está em WebhookItem separado!
}

// WebhookItem - Item de webhook payload (salvo UMA VEZ por execução)
// PK: "WEBHOOK#PAYLOAD#<uuid>" - identifica o webhook único
// SK: "TIMESTAMP#<RFC3339>" - timestamp da execução
type WebhookItem struct {
	PK      string `dynamodbav:"PK"`      // "WEBHOOK#PAYLOAD#<uuid>"
	SK      string `dynamodbav:"SK"`      // "TIMESTAMP#<RFC3339>"
	User    string `dynamodbav:"user"`    // Nome do usuário
	Payload string `dynamodbav:"payload"` // JSON completo do webhook
}

// FalhaItem - Item de erro/falha com UUID
// PK: "ERROR#<uuid>" - identifica o erro único
// SK: "TIMESTAMP#<RFC3339>" - timestamp do erro
type FalhaItem struct {
	PK              string `dynamodbav:"PK"`              // "ERROR#<uuid>"
	SK              string `dynamodbav:"SK"`              // "TIMESTAMP#<RFC3339>"
	ErrorType       string `dynamodbav:"errorType"`       // Tipo do erro
	ErrorMessage    string `dynamodbav:"errorMessage"`    // Mensagem do erro
	OriginalPayload string `dynamodbav:"originalPayload"` // Payload que causou o erro
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

// User locations response structure
type UserLocation struct {
	User      string `json:"user"`
	AvatarURL string `json:"avatarURL"`
	ISO3      string `json:"iso3"`
	Pais      string `json:"pais"`
	Livro     string `json:"livro"`
	Timestamp string `json:"timestamp"`
}

type UserLocationsResponse struct {
	Users []UserLocation `json:"users"`
	Total int            `json:"total"`
}
