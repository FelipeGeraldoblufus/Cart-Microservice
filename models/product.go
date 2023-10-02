package models

type Product struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null" json:"description"`
	Price       int    `gorm:"not null" json:"price"`
	Image       []byte `gorm:"not null" json:"image"`
}
