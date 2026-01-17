package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID             uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	ConversationID uuid.UUID      `gorm:"type:char(36);not null" json:"conversation_id"`
	SourceType     string         `gorm:"type:varchar(255);not null" json:"source_type"`
	SourceUri      string         `gorm:"type:varchar(255);not null" json:"source_uri"`
	CreatedAt      time.Time      `gorm:"type:timestamp;not null" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamp;not null" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Document) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
