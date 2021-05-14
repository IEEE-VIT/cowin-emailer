package db

type District struct {
	DistrictID   uint `gorm:"primaryKey"`
	DistrictName string
}
