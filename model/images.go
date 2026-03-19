package model

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	MimeType    string     `json:"mimeType" db:"mimeType"`
	Extension   string     `json:"extension" db:"extension"`
	Bytes       string     `json:"-" db:"bytes"`
	StoragePath string     `json:"storagePath" db:"storagePath"`
	Width       int        `json:"width" db:"width"`
	Height      int        `json:"height" db:"height"`
	CreatedAt   time.Time  `json:"createdAt" db:"createdAt"`
	GreyScaleID *uuid.UUID `json:"greyScaleId,omitempty" db:"greyScaleId"`
}

type PreviewImage struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OriginalImageID uuid.UUID `json:"originalImageId" db:"originalImageId"`
	StoragePath     string    `json:"storagePath" db:"storagePath"`
	Width           int       `json:"width" db:"width"`
	Height          int       `json:"height" db:"height"`
	CreatedAt       time.Time `json:"createdAt" db:"createdAt"`
	Extension       string    `json:"extension" db:"extension"`
	Name            string    `json:"name" db:"name"`
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
	OriginalImageId uuid.UUID `json:"originalImageId" db:"originalImageId"`
	StoragePath     string    `json:"storagePath" db:"storagePath"`
	Width           int       `json:"width" db:"width"`
	Height          int       `json:"height" db:"height"`
	CreatedAt       time.Time `json:"createdAt" db:"createdAt"`
}

type InsertGreyScaleImage struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OriginalImageId uuid.UUID `json:"originalImageId" db:"originalImageId"`
	StoragePath     string    `json:"storagePath" db:"storagePath"`
	CreatedAt       time.Time `json:"createdAt" db:"createdAt"`
}
