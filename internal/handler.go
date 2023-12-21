package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"

	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/controllers"
	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/models"
	"github.com/FelipeGeraldoblufus/Cart/controllers"
	"github.com/FelipeGeraldoblufus/Cart/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Handler(d amqp.Delivery, ch *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response models.Response
	log.Println(" [.] Received a message")

	var Payload struct {
		Pattern string          `json:"pattern"`
		Data    json.RawMessage `json:"data"`
		ID      string          `json:"id"`
	}
	var err error
	err = json.Unmarshal(d.Body, &Payload)

	actionType := Payload.Pattern

	//dataJSON, err := json.Marshal(Payload.Data)
	failOnError(err, "Failed to marshal data")
	switch actionType {
	case "GET_PRODUCT":
		log.Println(" [.] Getting products")
		/*products, err := controllers.GET()
		log.Println(products)
		failOnError(err, "Failed to get products")
		productsJSON, err := json.Marshal(products)
		failOnError(err, "Failed to marshal products")

		response = models.Response{
			Success: "succes",
			Message: "Products retrieved",
			Data:    productsJSON,
		}*/

	case "GET_USERBYNAME":
		log.Println(" [.] Getting product by Name")
		var data struct {
			Name string `json:"username"`
		}
		var err error
		var userJson []byte
		var users models.User

		err = json.Unmarshal(Payload.Data, &data)
		users, err = controllers.GetByUser(data.Name)

		userJson, err = json.Marshal(users)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error getting product",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "succes",
				Message: "Product retrieved",
				Data:    userJson,
			}
		}

	case "EDIT_PRODUCT":
		log.Println(" [.] Editing product by Name")
		var data struct {
			Product        string `json:"product"`
			NewNameProduct string `json:"newnameProduct"`
		}
		var err error
		var userJson []byte
		var producto models.Product

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data)

		// Decodificar los datos recibidos
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llamada a la función para actualizar el producto
		producto, err = controllers.UpdateProduct(data.Product, data.NewNameProduct)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating product",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Convertir el resultado a JSON y preparar la respuesta
		userJson, err = json.Marshal(producto)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Product updated",
				Data:    userJson,
			}
		}

	case "CREATE_PRODUCT":
		log.Println(" [.] Creating product")
		var data struct {
			Name string `json:"name"`
		}
		var err error
		var dataJson []byte
		var product models.Product
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Name)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		product, err = controllers.CreateProduct(data.Name)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating product",
				Data:    []byte(err.Error()),
			}
			break
		}
		dataJson, err = json.Marshal(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Product created",
				Data:    dataJson,
			}
		}

	case "DELETE_PRODUCT":
		log.Println(" [.] Deleting product")
		var data struct {
			Name string `json:"name"`
		}
		var err error
		var dataJson []byte
		var product models.Product
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Name)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		err = controllers.DeleteProductByName(data.Name)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error Deleting product",
				Data:    []byte(err.Error()),
			}
			break
		}
		dataJson, err = json.Marshal(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Product deleted",
				Data:    dataJson,
			}
		}

	case "CREATE_CARTITEM":
		log.Println(" [.] Creating cartitem")
		var data struct {
			Username    string `json:"username"`
			ProductName string `json:"productName"`
			Quantity    int    `json:"quantity"`
		}
		var err error
		var cartitem *models.CartItem // Cambiado a puntero

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username, data.ProductName, data.Quantity)
		var dataJson []byte
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Ahora asigna el resultado de la función a cartitem
		cartitem, err = controllers.AddCartItemToUserByID(data.Username, data.ProductName, data.Quantity)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Ahora puedes acceder a cartitem.ID
		dataJson, err = json.Marshal(cartitem.ID)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Cartitem created",
				Data:    dataJson,
			}
		}

	case "EDIT_USER":
		log.Println(" [.] Editing user")
		var data struct {
			CurrentUsername string `json:"currentUsername"`
			NewUsername     string `json:"newUsername"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CurrentUsername, data.NewUsername)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para editar el usuario
		_, err = controllers.EditUser(data.CurrentUsername, data.NewUsername)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error editing user",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User edited successfully",
			Data:    nil, // No necesitas enviar datos específicos en la respuesta
		}

	case "CREATE_USER":
		log.Println(" [.] Creating user")
		var data struct {
			Username string `json:"username"`
		}
		var err error

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Verificar que el campo necesario (username) no esté vacío
		if data.Username == "" {
			response = models.Response{
				Success: "error",
				Message: "Username is required",
				Data:    nil,
			}
			break
		}

		// Llama a la función para crear el usuario
		createdUser, err := controllers.CreateUser(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating user",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Convertir createdUser a formato JSON y luego a []byte
		userData, err := json.Marshal(createdUser)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error encoding user data",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User created successfully",
			Data:    userData,
		}

	case "DELETE_USER":
		log.Println(" [.] Deleting user")
		var data struct {
			Username string `json:"username"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.DeleteUser(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error deleting cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User deleted successfully",
			Data:    nil, // No necesitas enviar datos específicos en la respuesta
		}

	case "CREATE_ORDER":
		log.Println(" [.] Creating order")
		var data struct {
			Username    string `json:"username"`
			CartItemIDs []uint `json:"cartItemIDs"`
		}
		var err error

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Verificar que el campo necesario (username) no esté vacío
		if data.Username == "" {
			response = models.Response{
				Success: "error",
				Message: "Username is required",
				Data:    nil,
			}
			break
		}

		// Llama a la función para crear la orden
		createdOrder, err := controllers.CreateOrder(data.Username, data.CartItemIDs)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating order",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Convertir createdOrder a formato JSON y luego a []byte
		orderData, err := json.Marshal(createdOrder)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error encoding order data",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "Order created successfully",
			Data:    orderData,
		}
	case "GET_ORDERSBYUSERNAME":
		log.Println(" [.] Getting orders by Username")
		var data struct {
			Username string `json:"username"`
		}
		var err error
		var ordersJson []byte
		var orders []models.Order

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		orders, err = controllers.GetOrdersByUsername(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error getting orders",
				Data:    []byte(err.Error()),
			}
			break
		}

		ordersJson, err = json.Marshal(orders)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Orders retrieved",
				Data:    ordersJson,
			}
		}

	case "EDIT_CARTITEM":
		log.Println(" [.] updating cartitem")
		var data struct {
			CartItemID uint `json:"cartItemID"`
			Quantity   int  `json:"quantity"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CartItemID, data.Quantity)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.UpdateCartItemQuantity(data.CartItemID, data.Quantity)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem updated successfully",
			Data:    []byte("Cantidad actualizada exitosamente"), // No necesitas enviar datos específicos en la respuesta
		}

	case "EDIT_CARTITEMORDER":
		log.Println(" [.] updating cartitem")
		var data struct {
			CartItemID uint `json:"cartItemID"`
			Order      uint `json:"OrderID"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CartItemID, data.Order)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.UpdateCartItemOrder(data.CartItemID, data.Order)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem updated successfully",
			Data:    []byte("Orden asignada exitosamente"), // No necesitas enviar datos específicos en la respuesta
		}

	case "DELETE_CARTITEM":
		log.Println(" [.] Deleting cartitem")
		var data struct {
			Username   string `json:"username"`
			CartItemID uint   `json:"cartItemID"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username, data.CartItemID)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		_, err = controllers.RemoveCartItemFromUserByUsername(data.Username, data.CartItemID)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error deleting cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem deleted successfully",
			Data:    []byte("CartItem deleted successfully"), // No necesitas enviar datos específicos en la respuesta
		}
	case "CREATE_CATEGORY":
		log.Println(" [.] Creating category")
		//log.Println("data ", Payload.Data.Data)
		//log.Println("data JSON", dataJSON)

		/*var category models.Category
		err := json.Unmarshal(Payload.Data.Data, &category)
		failOnError(err, "Failed to unmarshal category")

		log.Println("category ", category)

		categoryJson, err := json.Marshal(category)
		failOnError(err, "Failed to marshal category")

		//err = json.Unmarshal(categoryJson, &category)

		_, err = controllers.CreateCategory(category)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating category",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "succes",
				Message: "Category created",
				Data:    categoryJson,
			}
		}*/

		/*case "GET_TOP3POPULARPRODUCTS":
		log.Println(" [.] Getting top 3 popular products")

		products, err := controllers.GetTop3PopularProducts()
		failOnError(err, "Failed to get products")
		productsJSON, err := json.Marshal(products)
		failOnError(err, "Failed to marshal products")

		response = models.Response{
			Success: "succes",
			Message: "Products retrieved",
			Data:    productsJSON,
		}*/
	}

	responseJSON, err := json.Marshal(response)
	failOnError(err, "Failed to marshal response")

	err = ch.PublishWithContext(ctx,
		"",        // exchange
		d.ReplyTo, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          responseJSON,
		})
	failOnError(err, "Failed to publish a message")

	d.Ack(false)
}
