package controllers

import (
	"log"

	db "github.com/FelipeGeraldoblufus/Cart/config"
	"github.com/FelipeGeraldoblufus/Cart/models"
)

func GetProducts() ([]models.Product, error) {
	var products []models.Product
	err := db.DB.Preload("Images").Select("id, name, price, brand, size_available, reviews, category_name").Order("name ASC").Find(&products).Error

	return products, err
}

func GetProductById(Productid string) (models.Product, error) {
	var product models.Product
	err := db.DB.Preload("Images").Select("id, name, description, price, brand, in_stock, size_available, reviews, category_name").Where("id = ?", Productid).First(&product).Error
	log.Println(product)
	return product, err
}

func CreateProduct(product models.Product) (models.Product, error) {
	err := db.DB.Create(&product).Error

	return product, err
}

func CreateCategory(category models.Category) (models.Category, error) {
	err := db.DB.Create(&category).Error

	return category, err
}

func GetTop3PopularProducts() ([]models.Product, error) {
	var products []models.Product
	/*err := db.DB.Model(&models.Visit{}).
	Select("products.name, products.price, products.brand, products.reviews, products.category_name, COUNT(visits.product_id) as visit_count").
	Select("array_agg(images.*) as images, COUNT(visits.product_id) as visit_count").
	Group("products.name, products.price, products.brand, products.reviews, products.category_name").
	Order("visit_count desc").
	Limit(3).
	Joins("JOIN products ON products.id = visits.product_id").
	Joins("LEFT JOIN product_images as images ON products.id = images.product_id").
	Find(&products).Error*/
	// Subconsulta para calcular el recuento de visitas por producto
	err := db.DB.Table("products").
		Select("products.id, products.name, products.price, products.brand, products.in_stock, products.size_available, products.reviews, products.category_name, COUNT(visits.product_id) as visit_count").
		Joins("LEFT JOIN visits ON products.id = visits.product_id").
		Group("products.id").
		Order("visit_count desc").
		Limit(3).
		Preload("Images").
		Find(&products).Error

	log.Println(products)
	return products, err
}
