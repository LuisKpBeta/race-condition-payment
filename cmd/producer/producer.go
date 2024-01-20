package main

import (
	"fmt"
	"os"
	"strconv"
	rc "user-payment/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	event, id := getComand()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	rc.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	rc.FailOnError(err, "Failed to open a channel")
	rc.ConnectProducer(ch)
	//run producer for consumer, sending Id equals to val
	if event == "c" {
		rc.ProduceClientMessage(ch, id)
	}
	if event == "p" {
		rc.ProducePaymentMessage(ch, id)
	}
}
func getComand() (string, int) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <p/c> <client_id>")
		return "", 0
	}
	arg := os.Args[1]
	value := os.Args[2]
	intValue, _ := strconv.Atoi(value)
	return arg, intValue
}
