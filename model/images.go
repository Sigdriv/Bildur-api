package model

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt"`
	MimeType    string    `json:"mimeType" db:"mimeType"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	StoragePath string    `json:"storagePath" db:"storagePath"`
	Bytes       string    `json:"-" db:"bytes"`
}

type PreviewImage struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ParentID    uuid.UUID `json:"parentId" db:"imageId"`
	VariantName string    `json:"variantName" db:"variantName"`
	StoragePath string    `json:"storagePath" db:"storagePath"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt"`
}
