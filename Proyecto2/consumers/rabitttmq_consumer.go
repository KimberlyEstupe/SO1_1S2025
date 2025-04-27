// consumers/rabbitmq_consumer.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

type WeatherTweet struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

func main() {
	// Configurar logger
	log.SetOutput(os.Stdout)
	log.Println("Iniciando consumidor de RabbitMQ...")

	// Conectar a RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:password@rabbitmq.message-broker.svc.cluster.local:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error al conectar con RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir canal: %v", err)
	}
	defer ch.Close()

	// Declarar cola
	q, err := ch.QueueDeclare(
		"weather_tweets", // nombre
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Error al declarar cola: %v", err)
	}

	// Configurar QoS para distribución justa de mensajes entre consumidores
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Error al configurar QoS: %v", err)
	}

	// Conectar a Valkey
	valkeyClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("VALKEY_ADDR"),
		Password: os.Getenv("VALKEY_PASSWORD"),
		DB:       0,
	})
	ctx := context.Background()

	// Verificar conexión a Valkey
	_, err = valkeyClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error al conectar con Valkey: %v", err)
	}
	defer valkeyClient.Close()

	// Consumir mensajes
	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Error al registrar consumidor: %v", err)
	}

	// Canal para señales de cierre
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// WaitGroup para esperar a que todas las goroutines terminen
	var wg sync.WaitGroup

	// Número de workers
	numWorkers := 10
	log.Printf("Iniciando %d workers para procesar mensajes", numWorkers)

	// Canal para distribuir mensajes a los workers
	workChan := make(chan amqp.Delivery, numWorkers)

	// Iniciar workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d iniciado", workerID)

			for msg := range workChan {
				processMessage(ctx, msg, valkeyClient)
				msg.Ack(false) // Confirmar mensaje procesado
			}

			log.Printf("Worker %d finalizado", workerID)
		}(i)
	}

	// Goroutine para recibir mensajes y enviarlos a los workers
	go func() {
		for msg := range messages {
			workChan <- msg
		}
	}()

	// Esperar señal de cierre
	<-stop
	log.Println("Cerrando consumidor...")

	// Cerrar canal de trabajo para que los workers terminen
	close(workChan)

	// Esperar a que todos los workers terminen
	wg.Wait()
	log.Println("Consumidor cerrado correctamente")
}

func processMessage(ctx context.Context, msg amqp.Delivery, valkeyClient *redis.Client) {
	// Decodificar mensaje
	var tweet WeatherTweet
	if err := json.Unmarshal(msg.Body, &tweet); err != nil {
		log.Printf("Error al decodificar mensaje: %v", err)
		return
	}

	// Incrementar contador por país
	countryKey := "rabbit:country:" + tweet.Pais
	if err := valkeyClient.HIncrBy(ctx, "rabbit:stats", countryKey, 1).Err(); err != nil {
		log.Printf("Error al incrementar contador de país: %v", err)
	}

	// Incrementar contador por tipo de clima
	weatherKey := "rabbit:weather:" + tweet.Clima
	if err := valkeyClient.HIncrBy(ctx, "rabbit:stats", weatherKey, 1).Err(); err != nil {
		log.Printf("Error al incrementar contador de clima: %v", err)
	}

	// Incrementar contador total
	if err := valkeyClient.HIncrBy(ctx, "rabbit:stats", "total", 1).Err(); err != nil {
		log.Printf("Error al incrementar contador total: %v", err)
	}

	// Guardar último mensaje para cada país
	if err := valkeyClient.HSet(ctx, "rabbit:last_messages", tweet.Pais, tweet.Descripcion).Err(); err != nil {
		log.Printf("Error al guardar último mensaje: %v", err)
	}
}