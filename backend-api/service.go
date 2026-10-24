package main

import (
	"context"
	"database/sql"
	"fmt"
)

type vehicleService struct {
	db *sql.DB
}

func newVehicleService(db *sql.DB) *vehicleService {
	return &vehicleService{
		db: db,
	}
}

type vehicleLocation struct {
	VehicleId string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type vehicleLocationQuery struct {
	TimestampStart int64
	TimestampEnd   int64
	Limit          int
}

func (s *vehicleService) GetLatestLocation(ctx context.Context, vehicleId string) (*vehicleLocation, error) {
	items, err := s.getLocationsByVehicleId(ctx, vehicleId, &vehicleLocationQuery{
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

func (s *vehicleService) GetLocationsByRange(
	ctx context.Context,
	vehicleId string,
	timeStart, timeEnd int64,
) ([]*vehicleLocation, error) {
	items, err := s.getLocationsByVehicleId(ctx, vehicleId, &vehicleLocationQuery{
		TimestampStart: timeStart,
		TimestampEnd:   timeEnd,
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *vehicleService) getLocationsByVehicleId(
	ctx context.Context,
	vehicleId string,
	vq *vehicleLocationQuery,
) ([]*vehicleLocation, error) {
	datapipe := []any{vehicleId}
	q := fmt.Sprintf(`SELECT vehicle_id, latitude, longitude, "timestamp" FROM vehicle_locations
	      WHERE vehicle_id = $%d
		`, len(datapipe))

	if vq.TimestampStart > 0 && vq.TimestampEnd > 0 {
		datapipe = append(datapipe, vq.TimestampStart, vq.TimestampEnd)
		q += fmt.Sprintf(` AND "timestamp" BETWEEN $%d AND $%d`, len(datapipe)-1, len(datapipe))
	}

	q += ` ORDER BY "timestamp" DESC`

	if vq.Limit > 0 {
		datapipe = append(datapipe, vq.Limit)
		q += fmt.Sprintf(` LIMIT $%d`, len(datapipe))
	}

	rows, err := s.db.QueryContext(
		ctx,
		q,
		datapipe...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*vehicleLocation
	for rows.Next() {
		var vl vehicleLocation
		if err := rows.Scan(&vl.VehicleId, &vl.Latitude, &vl.Longitude, &vl.Timestamp); err != nil {
			return nil, err
		}
		res = append(res, &vl)
	}

	return res, nil
}
