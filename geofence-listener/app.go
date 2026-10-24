package main

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type application struct {
	config        *config
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	handler       *handler
}

func newApplication(cfg *config) (*application, error) {
	app := new(application)
	app.config = cfg

	if err := app.initRabbitmqClient(); err != nil {
		return nil, err
	}

	if err := app.setupHandler(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *application) initRabbitmqClient() error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		app.config.rabbitmq.username,
		app.config.rabbitmq.password,
		app.config.rabbitmq.host,
		app.config.rabbitmq.port,
	))
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	if err := ch.ExchangeDeclare(
		RABBITMQ_EXCHANGE,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		conn.Close()
		return err
	}

	if _, err := ch.QueueDeclare(
		RABBITMQ_QUEUE,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		conn.Close()
		return err
	}

	if err := ch.QueueBind(
		RABBITMQ_QUEUE,
		RABBITMQ_KEY,
		RABBITMQ_EXCHANGE,
		false,
		nil,
	); err != nil {
		conn.Close()
		return err
	}

	log.Println("rabbitmq connection established")
	app.rabbitConn = conn
	app.rabbitChannel = ch

	return nil
}

func (app *application) setupHandler() error {
	h := newHandler()

	app.handler = h

	return nil
}

func (app *application) start(ctx context.Context) error {
	log.Printf("subscribing to rabbitmq queue %s", RABBITMQ_QUEUE)

	deliveries, err := app.rabbitChannel.ConsumeWithContext(
		ctx,
		RABBITMQ_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range deliveries {
		app.handler.msgHandler(msg)

		if err := msg.Ack(false); err != nil {
			log.Printf("error while acking message: %v", err)
		}
	}

	return nil
}

func (app *application) shutdown(ctx context.Context) error {
	if err := app.rabbitChannel.Close(); err != nil {
		log.Printf("error while closing rabbitmq channel: %v", err)
	}

	if err := app.rabbitConn.Close(); err != nil {
		log.Printf("error while closing rabbitmq connection: %v", err)
	}

	log.Println("rabbitmq closed")

	return nil
}
