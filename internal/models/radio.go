package models

import (
	"time"
)

// Radio represents a radio station
type Radio struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UUID       string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Name       string    `json:"name" gorm:"index;not null"`
	StreamURL  string    `json:"stream_url" gorm:"not null"`
	Country    string    `json:"country" gorm:"index"`
	Language   string    `json:"language"`
	Tags       string    `json:"tags"`
	Bitrate    int       `json:"bitrate"`
	Favicon    string    `json:"favicon"`
	Homepage   string    `json:"homepage"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
