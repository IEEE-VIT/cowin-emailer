package db

import (
	"gorm.io/gorm"
)

type Slot struct {
	gorm.Model
	UID         string `gorm:"uniqueIndex"`
	MinAgeLimit int64
	Date        string
	Name        string
	Address     string
	Pincode     int64
	From        string
	To          string
	District    string
	State       string
	Users       []*User `gorm:"many2many:user_slots;"`
}
