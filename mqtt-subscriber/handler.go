package main

import (
	"context"
	"encoding/json/v2"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type handler struct {
	service *vehicleService
}

func newHandler(svc *vehicleService) *handler {
	return &handler{
		service: svc,
	}
}

func (h *handler) msgHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("received message on topic %s: %s", msg.Topic(), msg.Payload())

	defer msg.Ack()

	var vl vehicleLocation
	if err := json.Unmarshal(msg.Payload(), &vl); err != nil {
		log.Printf("error while unmarshaling payload: %v", err)
		return
	}

	if vl.VehicleId == "" || vl.Timestamp <= 0 {
		log.Printf("invalid payload supplied: %v", vl)
		return
	}

	if vl.Latitude < -90 || vl.Latitude > 90 {
		log.Printf("invalid latitude supplied: %v", vl.Latitude)
		return
	}
	if vl.Longitude < -180 || vl.Longitude > 180 {
		log.Printf("invalid longitude supplied: %v", vl.Longitude)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.service.StoreLocation(ctx, &vl); err != nil {
		log.Printf("error while storing vehicle location: %v", err)
		return
	}

	if err := h.service.CheckGeofence(ctx, &vl); err != nil {
		log.Printf("error while checking geofence location: %v", err)
		return
	}

	return
}
