package entity

import (
	"time"
)

type FileMetadata struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
