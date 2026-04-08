package domain

import "time"

type Artist struct {
	ID        uint      `Gorm:"primaryKey;autoIncrement"`
	Name      string    `Gorm:"size:255;not null;uniqueIndex"`
	Slug      string    `Gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time `Gorm:"autoCreateTime"`
}
