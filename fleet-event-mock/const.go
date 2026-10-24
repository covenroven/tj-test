package main

import "time"

// threshold point
const (
	GEOPOINT_LONGITUDE = 106.8456
	GEOPOINT_LATITUDE  = -6.2088
)

const (
	VEHICLE_ID              = "B1234XYZ"
	PUBLISH_INTERVAL        = 2 * time.Second
	VEHICLE_SPEED_KMH       = 30.0
	VEHICLE_SPEED_MPS       = VEHICLE_SPEED_KMH / 3.6
	VEHICLE_START_LONGITUDE = 106.84424458285491 // 150m westward from threshold
	VEHICLE_START_LATITUDE  = GEOPOINT_LATITUDE
)
