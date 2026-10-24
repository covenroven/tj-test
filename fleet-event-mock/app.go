package main

import (
	"context"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/lib/pq"
)

type application struct {
	config     *config
	mqttClient mqtt.Client
	job        *job
}

func newApplication(cfg *config) (*application, error) {
	app := new(application)
	app.config = cfg

	if err := app.initMQTTClient(); err != nil {
		return nil, err
	}

	if err := app.setupJob(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *application) initMQTTClient() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(app.config.mqtt.broker)
	opts.SetClientID("fleet-event-mock")
	opts.SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	app.mqttClient = client

	return nil
}

func (app *application) setupJob() error {
	j := newJob(app.mqttClient, app.config.mqtt.topic)

	app.job = j

	return nil
}

func (app *application) start(ctx context.Context) error {
	log.Printf("publishing vehicle location every %s", PUBLISH_INTERVAL)

	ticker := time.NewTicker(PUBLISH_INTERVAL)
	defer ticker.Stop()

	for {
		if err := app.job.dispatch(); err != nil {
			log.Printf("error while dispatching job: %v", err)
		}

		select {
		case <-ctx.Done():
			log.Println("stopping app")
			return nil
		case <-ticker.C:
		}
	}
}

func (app *application) shutdown(ctx context.Context) error {
	app.mqttClient.Disconnect(2000)
	log.Println("subscriber closed")

	return nil
}
