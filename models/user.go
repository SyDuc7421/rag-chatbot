package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	FullName      string         `gorm:"size:255;not null" json:"full_name"`
	Email         string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"size:255;not null" json:"-"`
	Role          Role           `gorm:"type:enum('user','admin');default:'user'" json:"role"`
	Conversations []Conversation `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"conversations,omitempty"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
