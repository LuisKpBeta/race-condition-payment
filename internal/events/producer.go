package rabbitclient

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"
	"user-payment/internal/data"

	amqp "github.com/rabbitmq/amqp091-go"
)

var CLIENT_EXCHANGE = "clients"
var PAYMENTS_EXCHANGE = "payments"

func ConnectProducer(ch *amqp.Channel) {
	//create/connect to clients exchange
	err := ch.ExchangeDeclare(
		CLIENT_EXCHANGE, // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,
	)
	FailOnError(err, "Failed to declare an exchange for clients")

	//create/connect to payments exchange
	err = ch.ExchangeDeclare(
		PAYMENTS_EXCHANGE,   // name
		"x-delayed-message", // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		amqp.Table{ // arguments
			"x-delayed-type": "direct",
		},
	)
	FailOnError(err, "Failed to declare an exchange")
}

func ProduceClientMessage(ch *amqp.Channel, id int) {
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := data.Client{
		Id:   id,
		Name: "novo client",
	}
	body, _ := json.Marshal(client)
	err := ch.PublishWithContext(ctx,
		CLIENT_EXCHANGE, // exchange
		"",              // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Headers: amqp.Table{
				"x-delay": 5000,
				// 	"x-retry":       0,
				// 	"x-retry-limit": 10,
			},
			Body: []byte(body),
		})
	FailOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent Cient data %s", body)
}
func ProducePaymentMessage(ch *amqp.Channel, clientId int) {
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payment := data.Payment{
		Id:            generateId(),
		ClientId:      clientId,
		OperationDate: time.Now(),
		Value:         200,
	}
	body, _ := json.Marshal(payment)
	err := ch.PublishWithContext(ctx,
		PAYMENTS_EXCHANGE, // exchange
		"process.payment", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Headers: amqp.Table{
				"x-delay":       0,
				"x-retry":       0,
				"x-retry-limit": 10,
			},
			Body: []byte(body),
		})
	FailOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent Payment data %s", body)
}

func SendPaymentForReprocessQueue(ch *amqp.Channel, d amqp.Delivery) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	actualDelay := d.Headers["x-delay"].(int32)
	newDelay := actualDelay + 1000
	actualRetry, _ := d.Headers["x-retry"].(int32)
	newRetry := actualRetry + 1

	err := ch.PublishWithContext(ctx,
		PAYMENTS_EXCHANGE,   // exchange
		"reprocess.payment", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Headers: amqp.Table{
				"x-delay": newDelay,
				"x-retry": newRetry,
				// "x-retry-limit": 10,
			},
			Body: d.Body,
		})
	FailOnError(err, "Failed to publish a message")
	return err
}
func generateId() int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return r.Int()

}
