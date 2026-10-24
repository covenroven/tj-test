package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type config struct {
	rabbitmq *rabbitmqConfig
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
	grp, grpCtx := errgroup.WithContext(ctx)

	// start the main consumer; it blocks until the context is cancelled
	grp.Go(func() error {
		return app.start(grpCtx)
	})

	// block the program until the consumer stops
	runErr := grp.Wait()

	if err := app.shutdown(ctx); err != nil {
		log.Printf("error while shutting down: %v", err)
	}

	if runErr != nil {
		log.Fatalf("listener stopped with error: %v", runErr)
	}

	log.Println("app has been shutdown")
}

func loadConfig() *config {
	return &config{
		rabbitmq: &rabbitmqConfig{
			host:     getEnv("RABBITMQ_HOST", "localhost"),
			port:     getEnvInt("RABBITMQ_PORT", 5672),
			username: getEnv("RABBITMQ_USER", "rabbit"),
			password: getEnv("RABBITMQ_PASSWORD", "pass"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid %s=%q, using default %d", key, value, fallback)
		return fallback
	}

	return parsed
}
