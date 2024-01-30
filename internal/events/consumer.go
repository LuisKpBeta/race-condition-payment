package rabbitclient

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"user-payment/internal/data"

	amqp "github.com/rabbitmq/amqp091-go"
)

var CLIENT_PROCESSOR_QUEUE = "clients-processor"
var PAYMENT_PROCESSOR_QUEUE = "payment-processor"
var PAYMENT_REPROCESSOR_QUEUE = "payment-reprocessor"

func CreateQueue(ch *amqp.Channel, queueName string) amqp.Queue {
	mainQueue, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare a payment queue")
	return mainQueue
}
func CreateConsumer(ch *amqp.Channel, queue amqp.Queue, exchangeName, routingKey, consumerName string) <-chan amqp.Delivery {
	err := ch.QueueBind(
		queue.Name,   // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	msg := fmt.Sprintf("Failed to bind on %s queue using routingKey %s \n", exchangeName, routingKey)
	FailOnError(err, msg)
	msgs, err := ch.Consume(
		queue.Name,   // queue
		consumerName, // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	msg = fmt.Sprintf("Failed to start consumer %s on exchange %s \n", consumerName, exchangeName)
	FailOnError(err, msg)
	return msgs
}

func ConnectClientReceiver(ch *amqp.Channel, queue amqp.Queue, db *sql.DB) {
	msgsChannel := CreateConsumer(ch, queue, CLIENT_EXCHANGE, "", "client-consumer")
	for d := range msgsChannel {
		log.Println("processing message on client consumer")
		var client data.Client
		err := json.Unmarshal(d.Body, &client)
		if err != nil {
			log.Println("invalid message: ", d.Body)
			continue
		}
		data.AddClient(db, client)
	}
}
func ConnectPaymentReceiver(ch *amqp.Channel, queue amqp.Queue, db *sql.DB) {
	msgsChannel := CreateConsumer(ch, queue, PAYMENTS_EXCHANGE, "process.payment", "payment-consumer")

	for d := range msgsChannel {
		log.Println("processing message on payment consumer")
		validateAndProcessPayment(ch, db, d)
	}
}

func ConnectPaymentReProcessorReceiver(ch *amqp.Channel, queue amqp.Queue, db *sql.DB) {
	msgsChannel := CreateConsumer(ch, queue, PAYMENTS_EXCHANGE, "reprocess.payment", "payment-consumer-reprocess")
	for d := range msgsChannel {
		log.Println("receiving a payment to reprocess")
		validateAndProcessPayment(ch, db, d)
	}
}

func validateAndProcessPayment(ch *amqp.Channel, db *sql.DB, message amqp.Delivery) error {
	var payment data.Payment
	err := json.Unmarshal(message.Body, &payment)
	if err != nil {
		log.Println("invalid message: ", message.Body)
		return err
	}
	cli, _ := data.FindClientById(db, payment.ClientId)
	if cli.Id == 0 {
		SendPaymentForReprocessQueue(ch, message)
		log.Println("client dont exists, skipping payment ", payment.Id, " for client ", payment.ClientId)
		return nil
	}
	log.Println("processing payment id:", payment.Id, "for ", payment.ClientId)
	return nil
}
