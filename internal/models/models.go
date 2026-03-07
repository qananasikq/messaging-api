package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type DialogType string

const (
	DialogTypeDirect DialogType = "direct"
	DialogTypeGroup  DialogType = "group"
)

type Dialog struct {
	ID           uuid.UUID  `json:"id"`
	Type         DialogType `json:"type"`
	Name         *string    `json:"name,omitempty"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	Participants []UserMini `json:"participants,omitempty"`

	LastMessage *MessageMini `json:"last_message,omitempty"`
	UnreadCount int64        `json:"unread_count,omitempty"`
}

type UserMini struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	DialogID  uuid.UUID `json:"dialog_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageMini struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
