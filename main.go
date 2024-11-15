package main

import (
	"fmt"
	"log"

	"github.com/FelipeGeraldoblufus/product-microservice-go/config"
	"github.com/FelipeGeraldoblufus/product-microservice-go/internal"

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

// Declara la cola si no existe
func declareQueue(ch *amqp.Channel) amqp.Queue {
	// Declarar la cola si no existe
	q, err := ch.QueueDeclare(
		"product", // name
		true,      // durable (la cola sobrevivirá a reinicios de RabbitMQ)
		false,     // delete when unused (no se elimina automáticamente cuando está vacía)
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
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

	fmt.Println("Product MS starting...")

	// Cargar las variables de entorno
	godotenv.Load()
	fmt.Println("Loaded env variables...")

	// Configurar la base de datos
	config.SetupDatabase()
	fmt.Println("Database connection configured...")

	// Configurar RabbitMQ
	config.SetupRabbitMQ()
	fmt.Println("RabbitMQ Connection configured...")

	// Obtener canal de RabbitMQ
	ch := getChannel()
	// Declarar la cola y obtener su estructura
	q := declareQueue(ch)
	// Establecer la calidad de servicio en el canal
	setQoS(ch)
	// Registrar un consumidor para la cola
	msgs := registerConsumer(ch, q)

	// Iniciar el procesamiento de mensajes en un goroutine
	var forever chan struct{}
	go func() {
		for d := range msgs {
			// Llamar al manejador de mensajes internos con el mensaje y el canal de RabbitMQ
			internal.Handler(d, ch)
		}
	}()

	// Esperar indefinidamente a los mensajes
	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}
