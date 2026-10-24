package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type application struct {
	config        *config
	mqttClient    mqtt.Client
	rabbitChannel *amqp.Channel
	db            *sql.DB
	handler       *handler
}

func newApplication(cfg *config) (*application, error) {
	app := new(application)
	app.config = cfg

	if err := app.initMQTTClient(); err != nil {
		return nil, err
	}

	if err := app.initPostgres(); err != nil {
		return nil, err
	}

	if err := app.initRabbitmqClient(); err != nil {
		return nil, err
	}

	if err := app.setupHandler(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *application) initMQTTClient() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(app.config.mqtt.broker)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	app.mqttClient = client

	return nil
}

func (app *application) initPostgres() error {
	conf := app.config.database

	db, err := sql.Open("postgres", conf.dsn())
	if err != nil {
		return err
	}

	log.Println("postgres connection established")

	app.db = db

	return nil
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
	app.rabbitChannel = ch

	return nil
}

func (app *application) setupHandler() error {
	repo := newVehicleService(app.db, app.rabbitChannel)
	h := newHandler(repo)

	app.handler = h

	return nil
}

func (app *application) start(ctx context.Context) error {
	log.Println("subscribing to mqtt topic")
	token := app.mqttClient.Subscribe(app.config.mqtt.topic, byte(0), app.handler.msgHandler)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (app *application) shutdown(ctx context.Context) error {
	app.mqttClient.Disconnect(2000)
	log.Println("subscriber closed")

	if err := app.db.Close(); err != nil {
		return err
	}
	log.Println("postgres closed")

	return nil
}
