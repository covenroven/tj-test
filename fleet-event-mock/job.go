package main

import (
	"encoding/json"
	"log"
	"math"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const metersPerDegree = 111_320.0

type job struct {
	client mqtt.Client
	topic  string
	pos    location
}

func newJob(client mqtt.Client, topic string) *job {
	return &job{
		client: client,
		topic:  topic,
		pos: location{
			Latitude:  VEHICLE_START_LATITUDE,
			Longitude: VEHICLE_START_LONGITUDE,
		},
	}
}

func (j *job) dispatch() error {
	j.advance(VEHICLE_SPEED_MPS * PUBLISH_INTERVAL.Seconds())

	vl := vehicleLocation{
		VehicleId: VEHICLE_ID,
		Latitude:  j.pos.Latitude,
		Longitude: j.pos.Longitude,
		Timestamp: time.Now().Unix(),
	}

	payload, err := json.Marshal(vl)
	if err != nil {
		return err
	}

	token := j.client.Publish(j.topic, 0, false, payload)
	if err := token.Error(); err != nil {
		return err
	}

	log.Printf("successfully published location to topic %s | vehicle=%s lat=%.6f lon=%.6f",
		j.topic,
		vl.VehicleId,
		vl.Latitude,
		vl.Longitude,
	)

	return nil
}

func (j *job) advance(meters float64) {
	// advance east
	j.pos.Longitude += meters / (metersPerDegree * math.Cos(j.pos.Latitude*math.Pi/180))
}

type vehicleLocation struct {
	VehicleId string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
