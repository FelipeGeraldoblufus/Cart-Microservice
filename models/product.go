package models

type Product struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null" json:"description"`
	Price       int    `gorm:"not null" json:"price"`
	Image       string `gorm:"not null" json:"image"`

	// Agrega un campo que represente la relación con la categoría
	CategoryID uint // Puedes usar uint o el tipo de dato que sea adecuado
	Category   Category
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null" json:"name"`

	// Agrega un campo que represente la relación inversa con los productos
	Products []Product
}
