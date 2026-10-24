package main

const (
	RABBITMQ_EXCHANGE = "fleet.events"
	RABBITMQ_QUEUE    = "geofence_alerts"
	RABBITMQ_KEY      = "alert"
)

const (
	DISTANCE_THRESHOLD_M = 50.0
	GEOPOINT_LONGITUDE   = 106.8456
	GEOPOINT_LATITUDE    = -6.2088
)
