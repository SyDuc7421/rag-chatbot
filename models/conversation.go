package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Conversation struct {
	ID        uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	UserID    uuid.UUID      `gorm:"type:char(36);index;not null" json:"user_id"`
	Messages  []Message      `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"messages,omitempty"`
	Documents []Document     `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE;" json:"documents,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
