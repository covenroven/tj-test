# fleet-location

## Requirements
- Docker & Docker Compose

## How to run
- Make .env from .env.example if environtment variables change is needed
- Run `docker compose up`

## Notes
This repository contains several subdirectories:
- `backend-api`: Backend service providing API to get vehicle locations
- `fleet-event-mock`: Script to publish vehicle location mock data to MQTT every 2 seconds
- `geofence-listener`: RabbitMQ subscriber that logs geofence events published by `mqtt-subscriber`
- `mqtt-subscriber`: eclipse-mosquitto subscriber that stores vehicle location event and publish geofence event if necessary
- `db`: Migration script which will autorun on initial docker compose start
- `mosquitto`: Configuration file for eclipse-mosquitto
