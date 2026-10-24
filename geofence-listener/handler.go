package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type handler struct {
}

func newHandler() *handler {
	return &handler{}
}

func (h *handler) msgHandler(msg amqp.Delivery) {
	log.Printf("received geofence event from %s: %s", msg.Exchange, msg.Body)

	return
}
