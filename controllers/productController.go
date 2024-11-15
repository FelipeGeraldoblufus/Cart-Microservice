package controllers

import (
	"errors"
	db "github.com/FelipeGeraldoblufus/product-microservice-go/config"
	"github.com/FelipeGeraldoblufus/product-microservice-go/models"
	"gorm.io/gorm"
)

func CreateUser(username string) (*models.User, error) {
	// Crear un nuevo usuario sin el carrito (carrito ha sido eliminado)
	newUser := models.User{
		Username: username,
	}

	// Verificar si el nombre de usuario ya existe en la base de datos
	var existingUser models.User
	if err := db.DB.Where("username = ?", newUser.Username).First(&existingUser).Error; err == nil {
		// Si el usuario ya existe, devolver un error
		return nil, errors.New("username already exists")
	}

	// Guardar el nuevo usuario en la base de datos
	if err := db.DB.Save(&newUser).Error; err != nil {
		// Si ocurre un error al guardar, devolverlo
		return nil, err
	}

	// Devolver el usuario creado
	return &newUser, nil
}

func GetUser(usuario string) ([]models.User, error) {
	var user []models.User
	err := db.DB.Find(&user).Error

	return user, err
}

func GetByUser(username string) (models.User, error) {
	var users models.User
	err := db.DB.Where("username = ?", username).Find(&users).Error

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

// CreateProduct crea un nuevo producto con el nombre proporcionado
// Si el producto ya existe, devuelve un error.
func CreateProduct(nameProduct string) (models.Product, error) {
	// Verificar si el producto ya existe en la base de datos
	var existingProduct models.Product
	if err := db.DB.Where("name = ?", nameProduct).First(&existingProduct).Error; err == nil {
		// Si el producto ya existe, devolver un error
		return models.Product{}, errors.New("product with the same name already exists")
	}

	// Crear un nuevo producto
	newProduct := models.Product{
		Name: nameProduct,
	}

	// Iniciar una transacción
	tx := db.DB.Begin()

	// Manejo de errores de la transacción
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Intentar almacenar el nuevo producto en la base de datos
	if err := tx.Create(&newProduct).Error; err != nil {
		tx.Rollback() // Deshacer la transacción si hay un error
		return models.Product{}, err
	}

	// Confirmar la transacción si no hay errores
	tx.Commit()

	// Devolver el producto creado
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

func AddCartItemToUserByID(username string, productName string, quantity int) (*models.CartItem, error) {
	// Buscar al usuario por ID
	var user models.User
	user, err := GetByUser(username)
	if err != nil {
		return nil, err
	}

	if err := db.DB.Preload("Cart.Product").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// Buscar el producto por nombre
	var product models.Product
	if err := db.DB.Where("name = ?", productName).First(&product).Error; err != nil {
		// Si el producto no existe, créalo antes de agregar al carrito
		newProduct := models.Product{Name: productName}
		if err := db.DB.Create(&newProduct).Error; err != nil {
			return nil, fmt.Errorf("Error creating product: %v", err)
		}
		product = newProduct
	}

	// Crear un nuevo CartItem con la cantidad especificada
	cartItem := models.CartItem{
		ProductID: product.ID,
		Quantity:  quantity,
		UserID:    user.ID,
	}

	// Guardar el nuevo CartItem en la base de datos
	if err := db.DB.Create(&cartItem).Error; err != nil {
		return nil, fmt.Errorf("Error creating cart item: %v", err)
	}

	// Agregar el nuevo CartItem al carrito del usuario
	user.Cart = append(user.Cart, cartItem)

	// Guardar los cambios en el usuario
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("Error updating user: %v", err)
	}

	return &cartItem, nil
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
	if err := db.DB.Where("username = ?", currentUsername).First(&existingUser).Error; err != nil {
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


