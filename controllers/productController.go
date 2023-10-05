package controllers

import (
	db "github.com/ecommerce-proyecto-integrador/products-microservice/mod/config"
	"github.com/ecommerce-proyecto-integrador/products-microservice/mod/models"
)

func GetProducts() ([]models.Product, error) {
	var products []models.Product
	err := db.DB.Select("id, name, price, image").Find(&products).Error

	return products, err
}

func GetProductById(Productid string) (models.Product, error) {
	var product models.Product
	err := db.DB.Select("id, name, description, price, image").Where("id = ?", Productid).First(&product).Error

	return product, err
}

func CreateProduct(product models.Product) (models.Product, error) {
	err := db.DB.Create(&product).Error

	return product, err
}
