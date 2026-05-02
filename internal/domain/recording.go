package domain

import "time"

type RecordingStatus string

const (
	StatusPending  RecordingStatus = "pending"
	StatusApproved RecordingStatus = "approved"
	StatusRejected RecordingStatus = "rejected"
)

type Recording struct {
	ID           uint            `Gorm:"primaryKey;autoIncrement"`
	Title        string          `Gorm:"size:500;not null"`
	Slug         string          `Gorm:"size:500;not null;uniqueIndex"`
	Description  *string         `Gorm:"type:text"`
	ArtistID     *uint           `json:"-" Gorm:"index"`
	ArtistName   string          `json:"-" Gorm:"size:255;not null"`
	Artist       *Artist         `Gorm:"foreignKey:ArtistID"`
	ConcertDate  *time.Time      `Gorm:"type:date"`
	YoutubeID    *string         `Gorm:"size:20"`
	ExternalURL  *string         `Gorm:"type:text"`
	ThumbnailURL *string         `Gorm:"type:text"`
	Status       RecordingStatus `Gorm:"size:20;default:pending;index"`
	SubmittedAt  time.Time       `Gorm:"autoCreateTime"`
	ModeratedAt  *time.Time
	CreatedAt    time.Time `Gorm:"autoCreateTime"`
}
