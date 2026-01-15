package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	// 0: Bot 1: User
	Sender         bool           `gorm:"default:true" json:"sender"`
	ConversationID uuid.UUID      `gorm:"type:char(36);index;not null" json:"conversation_id"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	TokenCount     int            `gorm:"default:0" json:"token_count"`
	CreatedAt      time.Time      `gorm:"type:timestamp;not null" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamp;not null" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
