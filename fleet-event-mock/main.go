package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type config struct {
	mqtt *mqttConfig
}

type mqttConfig struct {
	broker string
	topic  string
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
			topic:  "/fleet/vehicle/" + VEHICLE_ID + "/location",
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
