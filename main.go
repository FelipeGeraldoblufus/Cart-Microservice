package main

import (
	"fmt"
	"log"
	"net/http"

	//"github.com/ValeHenriquez/example-rabbit-go/users-server/config"
	//"github.com/ValeHenriquez/example-rabbit-go/users-server/internal"
	"github.com/FelipeGeraldoblufus/Cart/config"
	"github.com/FelipeGeraldoblufus/Cart/controllers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func getChannel() *amqp.Channel {
	ch := config.GetChannel()
	if ch == nil {
		log.Panic("Failed to get channel")
	}
	return ch
}

func declareQueue(ch *amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare(
		"products_queue", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q
}

// Establece la calidad de servicio (QoS) para el canal de RabbitMQ.
func setQoS(ch *amqp.Channel) {
	err := ch.Qos(
		1,     // prefetch count: Especifica cuántos mensajes puede recibir un consumidor antes de que se detenga la entrega. En este caso, se establece en 1.
		0,     // prefetch size: No se usa en este caso, se establece como 0.
		false, // global: Indica si estas configuraciones de QoS se aplican a nivel de canal o a nivel de conexión. En este caso, es a nivel de canal (false).
	)

	failOnError(err, "Failed to set QoS")
}

// Registra un consumidor para la cola dada y devuelve un canal de entrega de mensajes.
func registerConsumer(ch *amqp.Channel, q amqp.Queue) <-chan amqp.Delivery {
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	return msgs
}

func main() {

	fmt.Println("Users MS starting...")

	godotenv.Load()
	fmt.Println("Loaded env variables...")

	config.SetupDatabase()
	fmt.Println("Database connection configured...")

	//config.SetupRabbitMQ()
	fmt.Println("RabbitMQ Connection configured...")

	/*ch := getChannel()              // Obtiene un canal de RabbitMQ
	q := declareQueue(ch)           // Declara una cola y obtiene su estructura
	setQoS(ch)                      // Establece la calidad de servicio en el canal
	msgs := registerConsumer(ch, q) // Registra un consumidor para la cola y obtiene un canal de entrega de mensajes

	var forever chan struct{}
	go func() {
		for d := range msgs {
			internal.Handler(d, ch) // Llama al manejador de mensajes internos con el mensaje y el canal de RabbitMQ
		}
	}()*/

	r := mux.NewRouter()

	r.HandleFunc("/api/product", controllers.CreateProductRest).Methods("POST")
	r.HandleFunc("/api/product/{name}", controllers.GetProductRest).Methods("GET")
	r.HandleFunc("/api/product/{name}", controllers.DeleteProductRest).Methods("DELETE")
	r.HandleFunc("/api/product/{name}", controllers.UpdateProductRest).Methods("PUT")

	r.HandleFunc("/api/cartitem", controllers.CreateCartItemRest).Methods("POST")
	r.HandleFunc("/api/cartitem/{id}", controllers.GetCartItemRest).Methods("GET")
	r.HandleFunc("/api/cartitem/{id}", controllers.DeleteCartItemRest).Methods("DELETE")
	r.HandleFunc("/api/cartitem/{id}", controllers.UpdateCartItemRest).Methods("PUT")

	r.HandleFunc("/api/user", controllers.CreateUserRest).Methods("POST")
	r.HandleFunc("/api/user/{username}", controllers.GetUserRest).Methods("GET")
	r.HandleFunc("/api/user/addcartitem", controllers.AddCartItemToUser).Methods("POST")
	r.HandleFunc("/api/user/removecartitem", controllers.RemoveCartItemFromUser).Methods("DELETE")
	r.HandleFunc("/api/user/edituser", controllers.EditUser).Methods("PUT")

	http.ListenAndServe(":3000", r)

	/*log.Printf(" [*] Awaiting RPC requests")
	<-forever // Espera indefinidamente*/
}
