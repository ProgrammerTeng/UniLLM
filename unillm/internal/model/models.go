package model

import "time"

// User represents a registered platform user.
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Name         string    `json:"name"`
	Role         string    `json:"role" gorm:"default:user"` // user, admin
	Balance      float64   `json:"balance" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKey represents a user's API key for accessing the proxy.
type APIKey struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	Name      string    `json:"name"`
	KeyHash   string    `json:"-" gorm:"uniqueIndex;not null"`
	KeyPrefix string    `json:"key_prefix" gorm:"not null"` // first 8 chars for display
	Scope     string    `json:"scope" gorm:"default:full"`  // full, readonly
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

// Provider represents an upstream AI provider (OpenAI, Anthropic, etc).
type Provider struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null"` // openai, anthropic, google, deepseek
	BaseURL   string    `json:"base_url" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
}

// ProviderKey is one API key for an upstream provider (supports key pool).
type ProviderKey struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	ProviderID int64     `json:"provider_id" gorm:"index;not null"`
	KeyValue   string    `json:"-" gorm:"not null"` // encrypted in DB
	Weight     int       `json:"weight" gorm:"default:1"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	RPM        int       `json:"rpm" gorm:"default:60"` // rate limit per minute
	CreatedAt  time.Time `json:"created_at"`
}

// ModelConfig maps a public model name to an upstream provider + model.
type ModelConfig struct {
	ID             int64   `json:"id" gorm:"primaryKey"`
	PublicName     string  `json:"public_name" gorm:"uniqueIndex;not null"` // e.g. "gpt-4o"
	ProviderID     int64   `json:"provider_id" gorm:"index;not null"`
	UpstreamModel  string  `json:"upstream_model" gorm:"not null"` // actual model name at provider
	InputPricePer1M  float64 `json:"input_price_per_1m"`
	OutputPricePer1M float64 `json:"output_price_per_1m"`
	IsActive       bool    `json:"is_active" gorm:"default:true"`
	MaxTokens      int     `json:"max_tokens" gorm:"default:4096"`
	SupportsStream bool    `json:"supports_stream" gorm:"default:true"`
	SupportsTools  bool    `json:"supports_tools" gorm:"default:false"`
	SupportsVision bool    `json:"supports_vision" gorm:"default:false"`
}

// UsageLog records each API call for billing and analytics.
type UsageLog struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	UserID           int64     `json:"user_id" gorm:"index;not null"`
	APIKeyID         int64     `json:"api_key_id" gorm:"index"`
	ModelName        string    `json:"model_name" gorm:"index;not null"`
	ProviderName     string    `json:"provider_name"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	Cost             float64   `json:"cost"`
	Latency          float64   `json:"latency"` // seconds
	Status           string    `json:"status"`   // ok, error, timeout
	HTTPStatus       int       `json:"http_status"`
	IsStream         bool      `json:"is_stream"`
	CreatedAt        time.Time `json:"created_at" gorm:"index"`
}
