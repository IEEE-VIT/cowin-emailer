package db

import "time"

type UserSlot struct {
	UserID    uint `gorm:"primaryKey"`
	SlotID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
