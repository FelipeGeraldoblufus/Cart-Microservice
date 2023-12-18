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
func addcartitem(user models.User, product models.Product, q int) (models.CartItem, error) {

	var cartItem models.CartItem
	cartItem.ProductID = product.ID
	cartItem.Quantity = q
	cartItem.UserID = user.ID

	err := db.DB.Preload("Product").Create(&cartItem).Error

	//err := db.DB.Create(&product).Error

	return cartItem, err
}

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

func AddCartItemToUserByID(username string, productName string, quantity int) error {
	// Buscar al usuario por ID
	var user models.User
	user, err := GetByUser(username)
	if err != nil {
		return err
	}

	if err := db.DB.Preload("Cart.Product").First(&user, user.ID).Error; err != nil {
		return fmt.Errorf("User not found: %v", err)
	}
	// Buscar el producto por nombre
	var product models.Product
	if err := db.DB.Where("name = ?", productName).First(&product).Error; err != nil {
		// Si el producto no existe, créalo antes de agregar al carrito
		newProduct := models.Product{Name: productName}
		if err := db.DB.Create(&newProduct).Error; err != nil {
			return fmt.Errorf("Error creating product: %v", err)
		}
		product = newProduct
	}

	// Buscar si el producto ya está en el carrito
	for _, cartItem := range user.Cart {
		if cartItem.ProductID == product.ID {
			// Actualizar la cantidad del producto si ya está en el carrito
			cartItem.Quantity += quantity
			if err := db.DB.Save(&cartItem).Error; err != nil {
				return fmt.Errorf("Error updating cart item: %v", err)
			}
			// Actualizar otras propiedades del usuario si es necesario.
			if err := db.DB.Save(&user).Error; err != nil {
				return fmt.Errorf("Error updating user: %v", err)
			}
			return nil
		}
	}

	// Si el producto no está en el carrito, agregar un nuevo CartItem
	var cartItem models.CartItem
	cartItem.ProductID = product.ID
	cartItem.Quantity = quantity
	cartItem.UserID = user.ID

	if err := db.DB.Preload("Product").Create(&cartItem).Error; err != nil {
		return fmt.Errorf("Error creating cart item: %v", err)
	}

	// Cargar manualmente la información del producto en el carrito
	for i, cartItem := range user.Cart {
		var product models.Product
		if err := db.DB.First(&product, cartItem.ProductID).Error; err != nil {
			return fmt.Errorf("Error loading product information: %v", err)
		}
		user.Cart[i].Product = product
	}

	// Actualizar otras propiedades del usuario si es necesario.
	user.Cart = append(user.Cart, cartItem)

	if err := db.DB.Save(&user).Error; err != nil {
		return fmt.Errorf("Error updating user: %v", err)
	}

	return nil
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
	fmt.Println("Request User ID:", requestData.UserID)
	var user models.User
	if err := db.DB.Preload("Cart.Product").First(&user, requestData.UserID).Error; err != nil {
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
	for _, cartItem := range user.Cart {
		if cartItem.ProductID == product.ID {
			// Actualizar la cantidad del producto si ya está en el carrito
			cartItem.Quantity += requestData.Quantity
			if err := db.DB.Save(&cartItem).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// Actualizar otras propiedades del usuario si es necesario.
			if err := db.DB.Save(&user).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// Devolver la respuesta actualizada
			json.NewEncoder(w).Encode(&user)
			return

		}
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

func RemoveCartItemFromUserByID(userID uint, cartItemID uint) (*models.User, error) {
	var user models.User
	if err := db.DB.Preload("Cart.Product").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// Buscar y eliminar el CartItem del carrito del usuario
	var cartItemToRemove models.CartItem
	for i, item := range user.Cart {
		if item.ID == cartItemID {
			cartItemToRemove = item

			// Paso 1: Eliminar el CartItem del carrito
			user.Cart = append(user.Cart[:i], user.Cart[i+1:]...)
			break
		}
	}

	// Verificar si se encontró el CartItem
	if cartItemToRemove.ID == 0 {
		return nil, fmt.Errorf("CartItem not found")
	}

	// Paso 2: Eliminar el CartItem de la base de datos
	if err := db.DB.Delete(&cartItemToRemove).Error; err != nil {
		return nil, fmt.Errorf("Error deleting CartItem: %v", err)
	}

	// Paso 3: Actualizar el usuario en la base de datos después de la eliminación
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("Error saving user: %v", err)
	}

	return &user, nil
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
	if err := db.DB.Preload("Cart.Product").First(&user, requestData.UserID).Error; err != nil {
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

func EditUserREST(w http.ResponseWriter, r *http.Request) {
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

func CreateOrder(username string, cartItemIDs []uint) (*models.Order, error) {
	user, err := GetByUser(username)
	if err != nil {
		return nil, err
	}

	// Preload de los CartItems para asegurar que se carguen con el usuario
	if err := db.DB.Preload("Cart.Product").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// Verificar si el carrito del usuario está vacío
	if len(user.Cart) == 0 {
		return nil, errors.New("cannot create order with an empty cart")
	}

	// Crear una nueva orden
	order := models.Order{
		UserID: user.ID,
		User:   user,
		Items:  make([]models.CartItem, 0), // Inicializar la lista de CartItems para evitar nil pointer
	}

	// Iterar sobre los CartItemIDs proporcionados
	for _, cartItemID := range cartItemIDs {
		// Buscar el CartItem en el carrito del usuario
		var cartItemToRemove *models.CartItem
		for i, item := range user.Cart {
			if item.ID == cartItemID {
				cartItemToRemove = &user.Cart[i]
				break
			}
		}

		// Verificar si se encontró el CartItem
		if cartItemToRemove == nil {
			return nil, fmt.Errorf("CartItem with ID %d not found in user's cart", cartItemID)
		}

		// Asignar el OrderID al CartItem
		cartItemToRemove.OrderID = order.ID

		// Actualizar el CartItem en la base de datos
		if err := db.DB.Save(cartItemToRemove).Error; err != nil {
			return nil, err
		}

		// Agregar el CartItem a la lista de la orden
		order.Items = append(order.Items, *cartItemToRemove)
	}

	// Limpiar el carrito del usuario después de crear la orden
	user.Cart = []models.CartItem{}

	// Guardar los cambios en el usuario
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	// Crear la orden
	if err := db.DB.Create(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func CreateOrderREST(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var user models.User
	if err := db.DB.Preload("Cart.Product").First(&user, "username = ?", requestData.Username).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	// Verificar si el carrito del usuario está vacío
	if len(user.Cart) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot create order with an empty cart"))
		return
	}

	// Crear una nueva orden
	order := models.Order{
		UserID: user.ID,
		User:   user,
		Items:  user.Cart,
	}

	if err := db.DB.Create(&order).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Limpiar el carrito del usuario después de crear la orden
	user.Cart = []models.CartItem{}

	for _, cartItem := range user.Cart {
		if err := db.DB.Delete(&cartItem).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	if err := db.DB.Save(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(&order)
}

func GetOrdersByUsername(username string) ([]models.Order, error) {
	var user models.User

	// Buscar al usuario por su nombre de usuario
	if err := db.DB.Preload("Cart.Product").Preload("Cart").Preload("Cart.Product").First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}

	var orders []models.Order

	// Buscar las órdenes asociadas al usuario
	if err := db.DB.Preload("User.Cart.Product").Preload("Items.Product").Find(&orders, "user_id = ?", user.ID).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func GetOrdersByUsernameREST(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var user models.User

	// Buscar al usuario por su nombre de usuario
	if err := db.DB.Preload("Cart.Product").Preload("Cart").Preload("Cart.Product").First(&user, "username = ?", params["username"]).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	var orders []models.Order

	// Buscar las órdenes asociadas al usuario
	if err := db.DB.Preload("User.Cart.Product").Preload("Items.Product").Find(&orders, "user_id = ?", user.ID).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Aquí 'user' tiene la información del usuario y 'orders' tiene la información de las órdenes.

	// Puedes enviar 'user' y 'orders' como respuesta JSON.

	json.NewEncoder(w).Encode(&orders)
}

func UpdateCartItemQuantity(cartItemID uint, newQuantity int) error {
	// Buscar el CartItem por su ID
	var cartItem models.CartItem
	if err := db.DB.First(&cartItem, cartItemID).Error; err != nil {
		return fmt.Errorf("CartItem not found: %v", err)
	}

	// Actualizar la cantidad del CartItem
	cartItem.Quantity = newQuantity

	// Guardar los cambios en la base de datos
	if err := db.DB.Save(&cartItem).Error; err != nil {
		return fmt.Errorf("Error updating CartItem quantity: %v", err)
	}

	return nil
}

func UpdateCartItemOrder(cartItemID uint, OrderID uint) error {
	// Buscar el CartItem por su ID
	var cartItem models.CartItem
	if err := db.DB.First(&cartItem, cartItemID).Error; err != nil {
		return fmt.Errorf("CartItem not found: %v", err)
	}

	// Actualizar la cantidad del CartItem
	cartItem.OrderID = OrderID

	// Guardar los cambios en la base de datos
	if err := db.DB.Save(&cartItem).Error; err != nil {
		return fmt.Errorf("Error updating CartItem quantity: %v", err)
	}

	return nil
}

func RemoveCartItemFromUserByUsername(username string, cartItemID uint) (*models.User, error) {
	// Obtener el usuario por nombre de usuario
	user, err := GetByUser(username)
	if err != nil {
		return nil, err
	}
	if err := db.DB.Preload("Cart.Product").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// Buscar y eliminar el CartItem del carrito del usuario
	var cartItemToRemove models.CartItem
	for i, item := range user.Cart {
		if item.ID == cartItemID {
			cartItemToRemove = item

			// Paso 1: Eliminar el CartItem del carrito
			user.Cart = append(user.Cart[:i], user.Cart[i+1:]...)
			break
		}
	}

	// Verificar si se encontró el CartItem
	if cartItemToRemove.ID == 0 {
		return nil, fmt.Errorf("CartItem not found")
	}

	// Paso 2: Eliminar el CartItem de la base de datos
	if err := db.DB.Delete(&cartItemToRemove).Error; err != nil {
		return nil, fmt.Errorf("Error deleting CartItem: %v", err)
	}

	// Paso 3: Actualizar el usuario en la base de datos después de la eliminación
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("Error saving user: %v", err)
	}

	return &user, nil
}
