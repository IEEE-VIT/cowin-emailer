package db

import "gorm.io/gorm"

type User struct {
	gorm.Model `json:"-"`
	UID        string  `json:"uid" gorm:"unique"`
	Email      string  `json:"email" gorm:"unique"`
	District   string  `json:"district"`
	Age        uint    `json:"age"`
	Slots      []*Slot `gorm:"many2many:user_slots;"`
}
