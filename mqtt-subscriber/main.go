package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type config struct {
	mqtt     *mqttConfig
	database *databaseConfig
	rabbitmq *rabbitmqConfig
}

type mqttConfig struct {
	broker string
	topic  string
}

type databaseConfig struct {
	host     string
	port     int
	dbName   string
	username string
	password string
}

func (c databaseConfig) dsn() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		c.host,
		c.port,
		c.dbName,
		c.username,
		c.password,
	)
}

type rabbitmqConfig struct {
	host     string
	port     int
	username string
	password string
	queue    string
	exchange string
}

func main() {
	cfg := loadConfig()

	app, err := newApplication(cfg)
	if err != nil {
		panic(err)
	}

	// make cancellable context for app graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// watch for termination signal and transmit via context if presents
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			os.Interrupt,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		<-c

		log.Println("app is shutting down...")

		cancel()
	}()

	// prepares the goroutines for starting application
	// errgroup is used to group multiple goroutine processes using same context
	grp, grpCtx := errgroup.WithContext(ctx)

	// start the main server
	grp.Go(func() error {
		return app.start(grpCtx)
	})

	// if app terminates, shutdown the server
	grp.Go(func() error {
		<-grpCtx.Done()

		return app.shutdown(grpCtx)
	})

	// block the program until termination signal is received
	if err := grp.Wait(); err != nil {
		log.Fatalf("error while trying to shutdown: %v", err)
	}

	log.Println("app has been shutdown")
}

func loadConfig() *config {
	return &config{
		mqtt: &mqttConfig{
			broker: getEnv("MQTT_BROKER", "tcp://localhost:1883"),
			topic:  getEnv("MQTT_TOPIC", "/fleet/vehicle/+/location"),
		},
		database: &databaseConfig{
			host:     getEnv("POSTGRES_HOST", "localhost"),
			port:     getEnvInt("POSTGRES_PORT", 5432),
			dbName:   getEnv("POSTGRES_DB", "transjakarta"),
			username: getEnv("POSTGRES_USER", "postgres"),
			password: getEnv("POSTGRES_PASSWORD", "pass"),
		},
		rabbitmq: &rabbitmqConfig{
			host:     getEnv("RABBITMQ_HOST", "localhost"),
			port:     getEnvInt("RABBITMQ_PORT", 5672),
			username: getEnv("RABBITMQ_USER", "rabbit"),
			password: getEnv("RABBITMQ_PASSWORD", "pass"),
		},
	}
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	return val
}

func getEnvInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def
	}

	return parsed
}
