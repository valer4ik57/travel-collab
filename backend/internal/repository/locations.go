package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"travel-collab/backend/internal/models"
)

func (s *Store) ListLocations(ctx context.Context, tripID string) ([]models.Location, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT l.id, l.trip_id, l.name, l.description,
		       ST_Y(l.coordinates::geometry) AS lat,
		       ST_X(l.coordinates::geometry) AS lng,
		       l.category, l.visit_at, l.route_id::text, r.title, r.route_date, l.route_order,
		       l.created_by, l.created_at
		FROM locations l
		LEFT JOIN trip_routes r ON r.id = l.route_id
		WHERE l.trip_id = $1
		ORDER BY
			CASE WHEN r.route_date IS NULL THEN 1 ELSE 0 END,
			r.route_date ASC,
			r.sort_order ASC,
			CASE WHEN l.route_order IS NULL THEN 1 ELSE 0 END,
			l.route_order ASC,
			CASE WHEN l.visit_at IS NULL THEN 1 ELSE 0 END,
			l.visit_at ASC,
			l.created_at ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]models.Location, 0)
	for rows.Next() {
		var l models.Location
		if err := rows.Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.VisitAt, &l.RouteID, &l.RouteTitle, &l.RouteDate, &l.RouteOrder, &l.CreatedBy, &l.CreatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

func (s *Store) CreateLocation(ctx context.Context, tripID, userID, name string, description *string, lat, lng float64, category string, visitAt *time.Time, routeID *string) (models.Location, error) {
	var l models.Location
	var routeParam any
	if routeID != nil {
		routeParam = *routeID
	}
	err := s.DB.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO locations (trip_id, name, description, coordinates, category, visit_at, route_id, route_order, created_by)
			VALUES (
				$1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6, $7, $8,
				CASE WHEN $8::uuid IS NULL THEN NULL ELSE COALESCE((SELECT MAX(route_order) + 1 FROM locations WHERE trip_id = $1 AND route_id = $8), 1) END,
				$9
			)
			RETURNING *
		)
		SELECT i.id, i.trip_id, i.name, i.description,
		       ST_Y(i.coordinates::geometry), ST_X(i.coordinates::geometry), i.category, i.visit_at,
		       i.route_id::text, r.title, r.route_date, i.route_order, i.created_by, i.created_at
		FROM inserted i
		LEFT JOIN trip_routes r ON r.id = i.route_id
	`, tripID, name, description, lng, lat, category, visitAt, routeParam, userID).Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.VisitAt, &l.RouteID, &l.RouteTitle, &l.RouteDate, &l.RouteOrder, &l.CreatedBy, &l.CreatedAt)
	return l, err
}

func (s *Store) UpdateLocation(ctx context.Context, tripID, locationID, name string, description *string, lat, lng float64, category string, visitAt *time.Time, routeID *string) (models.Location, error) {
	var l models.Location
	var routeParam any
	if routeID != nil {
		routeParam = *routeID
	}
	err := s.DB.QueryRow(ctx, `
		WITH old AS (
			SELECT route_id FROM locations WHERE trip_id = $1 AND id = $2
		), updated AS (
			UPDATE locations
			SET name = $3,
			    description = $4,
			    coordinates = ST_SetSRID(ST_MakePoint($5, $6), 4326),
			    category = $7,
			    visit_at = $8,
			    route_id = $9,
			    route_order = CASE
			        WHEN $9::uuid IS NULL THEN NULL
			        WHEN (SELECT route_id FROM old) IS DISTINCT FROM $9::uuid THEN COALESCE((SELECT MAX(route_order) + 1 FROM locations WHERE trip_id = $1 AND route_id = $9 AND id <> $2), 1)
			        ELSE route_order
			    END
			WHERE trip_id = $1 AND id = $2
			RETURNING *
		)
		SELECT u.id, u.trip_id, u.name, u.description,
		       ST_Y(u.coordinates::geometry), ST_X(u.coordinates::geometry), u.category, u.visit_at,
		       u.route_id::text, r.title, r.route_date, u.route_order, u.created_by, u.created_at
		FROM updated u
		LEFT JOIN trip_routes r ON r.id = u.route_id
	`, tripID, locationID, name, description, lng, lat, category, visitAt, routeParam).Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.VisitAt, &l.RouteID, &l.RouteTitle, &l.RouteDate, &l.RouteOrder, &l.CreatedBy, &l.CreatedAt)
	return l, err
}

func (s *Store) DeleteLocation(ctx context.Context, tripID, locationID string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM locations WHERE trip_id = $1 AND id = $2`, tripID, locationID)
	return err
}

func (s *Store) IsLocationInTrip(ctx context.Context, tripID, locationID string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM locations WHERE trip_id = $1 AND id = $2)
	`, tripID, locationID).Scan(&exists)
	return exists, err
}

func (s *Store) GetLocation(ctx context.Context, tripID, locationID string) (models.Location, error) {
	var l models.Location
	err := s.DB.QueryRow(ctx, `
		SELECT l.id, l.trip_id, l.name, l.description,
		       ST_Y(l.coordinates::geometry), ST_X(l.coordinates::geometry), l.category, l.visit_at,
		       l.route_id::text, r.title, r.route_date, l.route_order, l.created_by, l.created_at
		FROM locations l
		LEFT JOIN trip_routes r ON r.id = l.route_id
		WHERE l.trip_id = $1 AND l.id = $2
	`, tripID, locationID).Scan(&l.ID, &l.TripID, &l.Name, &l.Description, &l.Lat, &l.Lng, &l.Category, &l.VisitAt, &l.RouteID, &l.RouteTitle, &l.RouteDate, &l.RouteOrder, &l.CreatedBy, &l.CreatedAt)
	return l, err
}

func (s *Store) ReorderRouteLocations(ctx context.Context, tripID, routeID string, locationIDs []string) ([]models.Location, error) {
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for index, locationID := range locationIDs {
		ct, err := tx.Exec(ctx, `
			UPDATE locations
			SET route_order = $4
			WHERE trip_id = $1 AND route_id = $2 AND id = $3
		`, tripID, routeID, locationID, index+1)
		if err != nil {
			return nil, err
		}
		if ct.RowsAffected() == 0 {
			return nil, pgx.ErrNoRows
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.ListLocations(ctx, tripID)
}
