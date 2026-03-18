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
	Extension   string    `json:"extension" db:"extension"`
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
	Extension   string    `json:"extension" db:"extension"`
	Name        string    `json:"name" db:"name"`
}

type InsertImage struct {
	Name        string `json:"name" validate:"required"`
	MimeType    string `json:"mimeType" validate:"required"`
	Extension   string `json:"extension" validate:"required"`
	Bytes       int64  `json:"bytes" validate:"required"`
	StoragePath string `json:"storagePath" validate:"required"`
	Width       int    `json:"width" validate:"required"`
	Height      int    `json:"height" validate:"required"`
}

type InsertThumbnailImage struct {
	ParentID    uuid.UUID `json:"parentId" db:"imageId"`
	VariantName string    `json:"variantName" db:"variantName"`
	StoragePath string    `json:"storagePath" db:"storagePath"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt"`
}
