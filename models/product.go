package models

type Product struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null;unique" json:"name"`
}

type User struct {
	ID       uint       `gorm:"primaryKey" json:"id"`
	Username string     `gorm:"not null;unique" json:"username"`
}
