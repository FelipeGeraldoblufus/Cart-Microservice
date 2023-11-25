package models

import "time"

type Product struct {
	ID              uint           `gorm:"primaryKey"`
	Name            string         `gorm:"not null" json:"name"`
	publicationDate string         `gorm:"not null" json:"publicationDate"`
	author          int            `gorm:"not null" json:"author"`
	Price           int            `gorm:"not null" json:"price"`
	Description     string         `gorm:"not null" json:"description"`
	totalStock      uint           `gorm:"not null" json:"totalStock"`
	SizeAvailable   string         `gorm:"not null" json:"sizeAvailable"`
	Images          []ProductImage `gorm:"foreignKey:ProductID" json:"images"`
	Reviews         string         `gorm:"not null" json:"reviews"`

	// Agrega un campo que represente la relación con la categoría
	CategoryName string `gorm:"foreignKey:Name" json:"Category"` // Puedes usar uint o el tipo de dato que sea adecuado
	Category     Category
}

type Category struct {
	Name string `gorm:"primaryKey" json:"name"`
}

type CartItem struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

type Cart struct {
	User  User       `json:"user"`
	Items []CartItem `json:"items"`
}

type Order struct {
	User  User       `json:"user"`
	Items []CartItem `json:"items"`
}

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type ProductImage struct {
	ID        uint   `gorm:"primaryKey"`
	ProductID uint   `gorm:"not null" json:"-"`
	Color     string `json:"color"`
	ColorCode string `json:"colorCode"`
	Image     string `json:"image"`
}

type Visit struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null json:productid"`
	UserID    uint      `gorm:"not null json:userID"` // Si deseas rastrear visitas por usuario, puedes agregar un campo UserID
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`

	// Define una relación con el producto
	Product Product `gorm:"foreignKey:ProductID"`
}
