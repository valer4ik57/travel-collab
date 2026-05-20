package repository

import (
	"context"

	"travel-collab/backend/internal/models"
)

func (s *Store) ListLocations(ctx context.Context, tripID string) ([]models.Location, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, trip_id, name, description,
		       ST_Y(coordinates::geometry) AS lat,
		       ST_X(coordinates::geometry) AS lng,
		       category, created_by, created_at
		FROM locations
		WHERE trip_id = $1
		ORDER BY created_at ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]models.Location, 0)
	for rows.Next() {
		var l models.Location
		if err := rows.Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.CreatedBy, &l.CreatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

func (s *Store) CreateLocation(ctx context.Context, tripID, userID, name string, description *string, lat, lng float64, category string) (models.Location, error) {
	var l models.Location
	err := s.DB.QueryRow(ctx, `
		INSERT INTO locations (trip_id, name, description, coordinates, category, created_by)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7)
		RETURNING id, trip_id, name, description,
		          ST_Y(coordinates::geometry), ST_X(coordinates::geometry), category, created_by, created_at
	`, tripID, name, description, lng, lat, category, userID).Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.CreatedBy, &l.CreatedAt)
	return l, err
}

func (s *Store) UpdateLocation(ctx context.Context, tripID, locationID, name string, description *string, lat, lng float64, category string) (models.Location, error) {
	var l models.Location
	err := s.DB.QueryRow(ctx, `
		UPDATE locations
		SET name = $3,
		    description = $4,
		    coordinates = ST_SetSRID(ST_MakePoint($5, $6), 4326),
		    category = $7
		WHERE trip_id = $1 AND id = $2
		RETURNING id, trip_id, name, description,
		          ST_Y(coordinates::geometry), ST_X(coordinates::geometry), category, created_by, created_at
	`, tripID, locationID, name, description, lng, lat, category).Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.CreatedBy, &l.CreatedAt)
	return l, err
}

func (s *Store) DeleteLocation(ctx context.Context, tripID, locationID string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM locations WHERE trip_id = $1 AND id = $2`, tripID, locationID)
	return err
}
