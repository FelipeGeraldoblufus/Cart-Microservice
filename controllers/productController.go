package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/FelipeGeraldoblufus/Cart/config"
	"github.com/FelipeGeraldoblufus/Cart/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func CreateUser(username string) (*models.User, error) {
	// Crear un nuevo usuario con el nombre proporcionado
	newUser := models.User{
		Username: username,
		Cart:     []models.CartItem{},
	}

	// Verifica si el nombre de usuario ya existe en la base de datos
	var existingUser models.User
	if err := db.DB.Where("username = ?", newUser.Username).First(&existingUser).Error; err == nil {
		return nil, err
	}

	// Asocia el carrito vacío al usuario y créalo
	if err := db.DB.Save(&newUser).Error; err != nil {
		return nil, err
	}

	return &newUser, nil
}

func GetUser(usuario string) ([]models.User, error) {
	var user []models.User
	err := db.DB.Preload("Cart.Product").Find(&user).Error

	return user, err
}

func GetByUser(username string) (models.User, error) {
	var users models.User
	err := db.DB.Preload("Cart.Product").Where("username = ?", username).Find(&users).Error

	return users, err
}

func UpdateProduct(productoIngresado string, newnameProduct string) (models.Product, error) {
	// Inicia una transacción
	tx := db.DB.Begin()
	defer func() {
		// Recupera la transacción en caso de error y finaliza la función
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Consulta la base de datos para obtener el producto existente por su nombre
	var producto models.Product
	if err := tx.Where("name = ?", productoIngresado).First(&producto).Error; err != nil {
		tx.Rollback()
		return producto, err
	}

	// Verifica si el nombre está siendo cambiado y si existe otro producto con el mismo nombre
	if productoIngresado != newnameProduct {
		var duplicateProduct models.Product
		if err := tx.Where("name = ?", newnameProduct).First(&duplicateProduct).Error; err == nil {
			// Ya existe un producto con el nuevo nombre
			tx.Rollback()
			return producto, errors.New("product with the same name already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Otro error al buscar el producto duplicado
			tx.Rollback()
			return producto, err
		}
	}

	// Actualiza los campos del producto existente con los nuevos valores
	producto.Name = newnameProduct
	// Puedes agregar más campos según sea necesario

	// Guarda los cambios en la base de datos
	if err := tx.Save(&producto).Error; err != nil {
		// Ocurrió un error al guardar en la base de datos, realiza un rollback
		tx.Rollback()
		return producto, err
	}

	// Confirma la transacción
	tx.Commit()

	return producto, nil
}

func CreateProduct(nameProduct string) (models.Product, error) {
	// Crea un nuevo producto con el nombre proporcionado
	newProduct := models.Product{
		Name: nameProduct,
	}

	// Abre una transacción
	tx := db.DB.Begin()

	// Maneja los errores de la transacción
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Aquí deberías almacenar el producto en la base de datos o realizar otras operaciones necesarias
	if err := tx.Create(&newProduct).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return models.Product{}, err
	}

	// Confirma la transacción si no hay errores
	tx.Commit()

	return newProduct, nil
}

func DeleteProductByName(nameProduct string) error {
	// Abre una transacción
	tx := db.DB.Begin()

	// Maneja los errores de la transacción
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Busca el producto por nombre
	var product models.Product
	if err := tx.Where("name = ?", nameProduct).First(&product).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Elimina el producto
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Confirma la transacción si no hay errores
	tx.Commit()

	return nil
}

func EditUser(currentUsername string, newUsername string) (*models.User, error) {
	// Buscar el usuario actual en la base de datos
	var existingUser models.User
	if err := db.DB.Preload("Cart.Product").Where("username = ?", currentUsername).First(&existingUser).Error; err != nil {
		return nil, err
	}

	// Modificar el nombre de usuario
	existingUser.Username = newUsername

	// Guardar los cambios en la base de datos
	if err := db.DB.Save(&existingUser).Error; err != nil {
		return nil, err
	}

	// Devolver el usuario actualizado
	return &existingUser, nil
}

func DeleteUser(usuario string) error {
	// Abre una transacción
	tx := db.DB.Begin()

	// Maneja los errores de la transacción
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Busca el producto por nombre
	var user models.User
	if err := tx.Where("username = ?", usuario).First(&user).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Elimina el usuario
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Confirma la transacción si no hay errores
	tx.Commit()

	return nil
}


