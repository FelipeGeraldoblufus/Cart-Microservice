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

func GetProductRest(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	params := mux.Vars(r)
	db.DB.First(&product, "name = ?", params["name"])
	if product.Name != params["name"] {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("product not found"))
		return

	}
	json.NewEncoder(w).Encode(&product)

}

func CreateProductRest(w http.ResponseWriter, r *http.Request) {

	var product models.Product
	json.NewDecoder(r.Body).Decode(&product)
	createdproduct := db.DB.Create(&product)
	err := createdproduct.Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return

	}

	json.NewEncoder(w).Encode(&product)
}

func DeleteProductRest(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	params := mux.Vars(r)

	db.DB.First(&product, "name = ?", params["name"])
	if product.Name != params["name"] {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("product not found"))
		return
	}

	db.DB.Unscoped().Delete(&product)
	w.WriteHeader(http.StatusNoContent)
}

func UpdateProductRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productName := vars["name"]

	var updatedProduct models.Product
	err := json.NewDecoder(r.Body).Decode(&updatedProduct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Consulta la base de datos para obtener el producto existente por su nombre
	var existingProduct models.Product
	if err := db.DB.Where("name = ?", productName).First(&existingProduct).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("product not found"))
		return
	}

	// Verifica si el nombre está siendo cambiado y si existe otro producto con el mismo nombre
	if updatedProduct.Name != existingProduct.Name {
		var duplicateProduct models.Product
		if err := db.DB.Where("name = ?", updatedProduct.Name).First(&duplicateProduct).Error; err == nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("product with the same name already exists"))
			return
		}
	}

	// Actualiza los campos del producto existente con los nuevos valores
	existingProduct.Name = updatedProduct.Name
	// Puedes agregar más campos según sea necesario

	// Guarda los cambios en la base de datos
	if err := db.DB.Save(&existingProduct).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Responde con el producto actualizado
	json.NewEncoder(w).Encode(&existingProduct)
}

func CreateCartItemRest(w http.ResponseWriter, r *http.Request) {
	var cartitem models.CartItem
	err := json.NewDecoder(r.Body).Decode(&cartitem)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Verifica si el producto existe en la base de datos
	var existingProduct models.Product
	if err := db.DB.First(&existingProduct, cartitem.ProductID).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Product not found"))
		return
	}

	// Imprime información sobre el producto existente
	fmt.Printf("Existing Product: %+v\n", existingProduct)

	// Crea el CartItem en la base de datos
	createdCartItem := db.DB.Create(&cartitem)
	if err := createdCartItem.Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Recarga el CartItem con la relación Product completamente cargada
	if err := db.DB.Preload("Product").First(&cartitem, cartitem.ID).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Responde con el CartItem creado
	json.NewEncoder(w).Encode(&cartitem)
}

func GetCartItemRest(w http.ResponseWriter, r *http.Request) {
	var cartitem models.CartItem
	params := mux.Vars(r)

	// Parsea el ID de cartitem desde los parámetros de la ruta
	cartitemID, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid cartitem ID: %v", err)))
		return
	}

	// Antes de cargar el CartItem, carga explícitamente el producto relacionado
	if err := db.DB.Preload("Product").First(&cartitem, cartitemID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("CartItem not found: %v", err)))
		return
	}

	// Resto del código para responder con el CartItem, por ejemplo:
	json.NewEncoder(w).Encode(&cartitem)
}

func UpdateCartItemRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartItemID := vars["id"]

	var updatedCartItem models.CartItem
	err := json.NewDecoder(r.Body).Decode(&updatedCartItem)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Consulta la base de datos para obtener el carrito existente por su ID
	var existingCartItem models.CartItem
	if err := db.DB.Preload("Product").First(&existingCartItem, cartItemID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Cart item not found"))
		return
	}

	// Actualiza los campos del carrito existente con los nuevos valores
	existingCartItem.Quantity = updatedCartItem.Quantity
	// Puedes agregar más campos según sea necesario

	// Asegúrate de que solo se actualice el producto si se proporciona en la solicitud
	if updatedCartItem.Product.ID != 0 {
		// Actualiza solo el ID del producto, si es necesario
		existingCartItem.ProductID = updatedCartItem.Product.ID
		// Puedes agregar más campos según sea necesario
	}

	// Guarda los cambios en la base de datos
	if err := db.DB.Save(&existingCartItem).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Responde con el carrito actualizado
	json.NewEncoder(w).Encode(&existingCartItem)
}

func DeleteCartItemRest(w http.ResponseWriter, r *http.Request) {
	var cartitem models.CartItem
	params := mux.Vars(r)

	fmt.Println("ID from params:", params["id"])

	cartItemID, err := strconv.Atoi(params["id"])
	if err != nil {
		fmt.Println("Error converting to int:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid cartitem ID"))
		return
	}

	// Busca el CartItem por su ID
	if err := db.DB.First(&cartitem, cartItemID).Error; err != nil {
		// Maneja el caso en que no se encuentra el CartItem
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("CartItem not found"))
			return
		}

		// Maneja otros posibles errores de la base de datos
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Elimina el CartItem
	if err := db.DB.Delete(&cartitem).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Respuesta exitosa
	w.WriteHeader(http.StatusNoContent)
}

func GetUserRest(w http.ResponseWriter, r *http.Request) {
	var user models.User
	params := mux.Vars(r)

	if err := db.DB.Preload("Cart.Product").First(&user, "username = ?", params["username"]).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	json.NewEncoder(w).Encode(&user)
}
func CreateUserRest(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Verifica si el nombre de usuario ya existe en la base de datos
	var existingUser models.User
	if err := db.DB.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Username already exists"))
		return
	}

	// Asocia el carrito vacío al usuario y creando al usuario
	user.Cart = []models.CartItem{}
	if err := db.DB.Save(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Responde con el usuario creado (incluyendo el carrito vacío)
	json.NewEncoder(w).Encode(&user)
}

// Agregar un producto al carrito de un usuario
func AddCartItemToUser(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID      uint   `json:"userID"`
		ProductName string `json:"productName"`
		Quantity    int    `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var user models.User
	if err := db.DB.Preload("Cart").First(&user, requestData.UserID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	var product models.Product
	if err := db.DB.Where("name = ?", requestData.ProductName).First(&product).Error; err != nil {
		// Si el producto no existe, créalo antes de agregar al carrito
		newProduct := models.Product{Name: requestData.ProductName}
		if err := db.DB.Create(&newProduct).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		product = newProduct
	}

	var cartItem models.CartItem
	cartItem.ProductID = product.ID
	cartItem.Quantity = requestData.Quantity
	cartItem.UserID = requestData.UserID

	if err := db.DB.Preload("Product").Create(&cartItem).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Cargar manualmente la información del producto en el carrito
	for i, cartItem := range user.Cart {
		var product models.Product
		if err := db.DB.First(&product, cartItem.ProductID).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		user.Cart[i].Product = product
	}

	// Actualizar otras propiedades del usuario si es necesario.
	user.Cart = append(user.Cart, cartItem)

	if err := db.DB.Save(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&user)
}

// RemoveCartItemFromUser elimina un elemento del carrito de un usuario
func RemoveCartItemFromUser(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID     uint `json:"userID"`
		CartItemID uint `json:"cartItemID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var user models.User
	if err := db.DB.Preload("Cart").First(&user, requestData.UserID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	// Buscar y eliminar el CartItem de la base de datos y del carrito del usuario
	var cartItemToRemove models.CartItem
	for i, item := range user.Cart {
		if item.ID == requestData.CartItemID {
			cartItemToRemove = item

			// Paso 1: Eliminar el CartItem de la base de datos
			if err := db.DB.Delete(&cartItemToRemove).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Paso 2: Eliminar el elemento del carrito
			user.Cart = append(user.Cart[:i], user.Cart[i+1:]...)
			break
		}
	}

	// Actualizar el usuario en la base de datos después de la eliminación
	if err := db.DB.Save(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&user)
}

func EditUser(w http.ResponseWriter, r *http.Request) {
	// Estructura para decodificar la solicitud
	var requestData struct {
		CurrentUsername string `json:"currentUsername"`
		NewUsername     string `json:"newUsername"`
	}

	// Decodificar la solicitud y obtener los nombres de usuario
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Buscar el usuario actual en la base de datos
	var existingUser models.User
	if err := db.DB.Preload("Cart.Product").Where("username = ?", requestData.CurrentUsername).First(&existingUser).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	// Modificar el nombre de usuario
	existingUser.Username = requestData.NewUsername

	// Guardar los cambios en la base de datos
	if err := db.DB.Save(&existingUser).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Responder con el usuario actualizado
	json.NewEncoder(w).Encode(&existingUser)
}

/*
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
}*/
/*
func GetTop3PopularProducts() ([]models.Product, error) {
	var products []models.Product
	err := db.DB.Model(&models.Visit{}).
	Select("products.name, products.price, products.brand, products.reviews, products.category_name, COUNT(visits.product_id) as visit_count").
	Select("array_agg(images.*) as images, COUNT(visits.product_id) as visit_count").
	Group("products.name, products.price, products.brand, products.reviews, products.category_name").
	Order("visit_count desc").
	Limit(3).
	Joins("JOIN products ON products.id = visits.product_id").
	Joins("LEFT JOIN product_images as images ON products.id = images.product_id").
	Find(&products).Error*/
// Subconsulta para calcular el recuento de visitas por producto
/*err := db.DB.Table("products").
	Select("products.id, products.name, products.price, products.brand, products.in_stock, products.size_available, products.reviews, products.category_name, COUNT(visits.product_id) as visit_count").
	Joins("LEFT JOIN visits ON products.id = visits.product_id").
	Group("products.id").
	Order("visit_count desc").
	Limit(3).
	Preload("Images").
	Find(&products).Error

log.Println(products)
return products, err*/
