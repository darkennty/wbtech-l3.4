package model

import "time"

type Image struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	OriginalPath  string    `json:"original_path"`
	WatermarkPath string    `json:"watermark_path,omitempty"`
	ThumbPath     string    `json:"processed_path,omitempty"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
