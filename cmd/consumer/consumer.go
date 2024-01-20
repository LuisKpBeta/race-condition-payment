package main

import (
	dt "user-payment/internal/data"
	rc "user-payment/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	rc.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	rc.FailOnError(err, "Failed to open a channel")

	db := dt.ConnectToDabase()
	var forever chan struct{}
	clientQueue := rc.CreateQueue(ch, rc.CLIENT_PROCESSOR_QUEUE)
	go rc.ConnectClientReceiver(ch, clientQueue, db)

	paymentQueue := rc.CreateQueue(ch, rc.PAYMENT_PROCESSOR_QUEUE)
	go rc.ConnectPaymentReceiver(ch, paymentQueue, db)

	paymentReprocessQueue := rc.CreateQueue(ch, rc.PAYMENT_REPROCESSOR_QUEUE)
	go rc.ConnectPaymentReProcessorReceiver(ch, paymentReprocessQueue, db)

	<-forever
}
