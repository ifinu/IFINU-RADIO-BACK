package models

import (
	"time"
)

// Radio represents a radio station
type Radio struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UUID        string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"index;not null"`
	Description string    `json:"description"`
	StreamURL   string    `json:"stream_url" gorm:"not null"`
	Country     string    `json:"country" gorm:"index"`
	State       string    `json:"state"`
	City        string    `json:"city"`
	Language    string    `json:"language"`
	Tags        string    `json:"tags"`
	Bitrate     int       `json:"bitrate"`
	Favicon     string    `json:"favicon"`
	Homepage    string    `json:"homepage"`
	Active      bool      `json:"active" gorm:"default:true"`
	Listeners   int       `json:"listeners" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToFrontendDTO converts Radio model to frontend expected format
type RadioDTO struct {
	ID          string `json:"id"`
	Nome        string `json:"nome"`
	Descricao   string `json:"descricao"`
	URLStream   string `json:"url_stream"`
	LogoURL     string `json:"logo_url"`
	Genero      string `json:"genero"`
	Pais        string `json:"pais"`
	Cidade      string `json:"cidade,omitempty"`
	Estado      string `json:"estado,omitempty"`
	Idioma      string `json:"idioma"`
	Ativo       bool   `json:"ativo"`
	Ouvintes    int    `json:"ouvintes"`
	CriadoEm    string `json:"criado_em"`
	AtualizadoEm string `json:"atualizado_em"`
}

func (r *Radio) ToDTO() RadioDTO {
	return RadioDTO{
		ID:          r.UUID,
		Nome:        r.Name,
		Descricao:   r.Description,
		URLStream:   r.StreamURL,
		LogoURL:     r.Favicon,
		Genero:      r.Tags,
		Pais:        r.Country,
		Cidade:      r.City,
		Estado:      r.State,
		Idioma:      r.Language,
		Ativo:       r.Active,
		Ouvintes:    r.Listeners,
		CriadoEm:    r.CreatedAt.Format(time.RFC3339),
		AtualizadoEm: r.UpdatedAt.Format(time.RFC3339),
	}
}

// RadioBrowserStation represents external API response
type RadioBrowserStation struct {
	StationUUID string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	URLResolved string `json:"url_resolved"`
	Country     string `json:"country"`
	Language    string `json:"language"`
	Tags        string `json:"tags"`
	Bitrate     int    `json:"bitrate"`
	Favicon     string `json:"favicon"`
	Homepage    string `json:"homepage"`
}
