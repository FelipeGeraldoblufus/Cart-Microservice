package models

type Product struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null;unique" json:"name"`
}

type CartItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	UserID    uint    `gorm:"not null" json:"user_id"`
}

type Order struct {
	ID     uint       `gorm:"primaryKey" json:"id"`
	UserID uint       `gorm:"not null" json:"user_id"`
	User   User       `gorm:"foreignKey:ID" json:"user"`
	Items  []CartItem `gorm:"foreignKey:ID" json:"items"`
}

type User struct {
	ID       uint       `gorm:"primaryKey" json:"id"`
	Username string     `gorm:"not null;unique" json:"username"`
	Cart     []CartItem `gorm:"foreignKey:UserID" json:"cart"`
}
