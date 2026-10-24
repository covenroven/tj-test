package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math"

	amqp "github.com/rabbitmq/amqp091-go"
)

const earthRadiusM = 6371000.0

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

type vehicleGeofence struct {
	VehicleId string   `json:"vehicle_id"`
	Event     string   `json:"event"`
	Location  location `json:"location"`
	Timestamp int64    `json:"timestamp"`
}

type vehicleService struct {
	db            *sql.DB
	rabbitChannel *amqp.Channel
}

func newVehicleService(db *sql.DB, ch *amqp.Channel) *vehicleService {
	return &vehicleService{
		db:            db,
		rabbitChannel: ch,
	}
}

func (s *vehicleService) StoreLocation(ctx context.Context, vl *vehicleLocation) error {
	q := `INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, "timestamp")
			VALUES ($1, $2, $3, $4)`

	if _, err := s.db.ExecContext(
		ctx,
		q,
		vl.VehicleId,
		vl.Latitude,
		vl.Longitude,
		vl.Timestamp,
	); err != nil {
		return err
	}

	return nil
}

func (s *vehicleService) CheckGeofence(ctx context.Context, vl *vehicleLocation) error {
	distance := distanceInMeters(
		vl.Latitude,
		vl.Longitude,
		GEOPOINT_LATITUDE,
		GEOPOINT_LONGITUDE,
	)

	if distance <= DISTANCE_THRESHOLD_M {
		log.Printf("vehicle %s is within the geofence (%.2fm)", vl.VehicleId, distance)

		if err := s.publishGeofenceEvent(ctx, vl); err != nil {
			log.Printf("error while publishing geofence event: %v", err)
		}
	} else {
		log.Printf("vehicle %s is %.2fm away from the geofence", vl.VehicleId, distance)
	}
	return nil
}

// calculate distance in meters using haversine formula
func distanceInMeters(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusM * c
}

func (s *vehicleService) publishGeofenceEvent(ctx context.Context, vl *vehicleLocation) error {
	vg := vehicleGeofence{
		VehicleId: vl.VehicleId,
		Event:     "geofence_entry",
		Location: location{
			Latitude:  vl.Latitude,
			Longitude: vl.Longitude,
		},
		Timestamp: vl.Timestamp,
	}

	payload, _ := json.Marshal(vg)
	if err := s.rabbitChannel.PublishWithContext(
		ctx,
		RABBITMQ_EXCHANGE,
		RABBITMQ_KEY,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Transient,
			Body:         payload,
		},
	); err != nil {
		return err
	}
	log.Println("successfully published geofence event")

	return nil
}
