package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"travel-collab/backend/internal/models"
)

func (s *Store) ListRoutes(ctx context.Context, tripID string) ([]models.TripRoute, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, trip_id, title, route_date, sort_order, created_at
		FROM trip_routes
		WHERE trip_id = $1
		ORDER BY
			CASE WHEN route_date IS NULL THEN 1 ELSE 0 END,
			route_date ASC,
			sort_order ASC,
			created_at ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := make([]models.TripRoute, 0)
	for rows.Next() {
		var route models.TripRoute
		if err := rows.Scan(&route.ID, &route.TripID, &route.Title, &route.RouteDate, &route.SortOrder, &route.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (s *Store) CreateRoute(ctx context.Context, tripID, title string, routeDate *time.Time) (models.TripRoute, error) {
	var route models.TripRoute
	err := s.DB.QueryRow(ctx, `
		INSERT INTO trip_routes (trip_id, title, route_date, sort_order)
		VALUES ($1, $2, $3, COALESCE((SELECT MAX(sort_order) + 1 FROM trip_routes WHERE trip_id = $1), 1))
		RETURNING id, trip_id, title, route_date, sort_order, created_at
	`, tripID, title, routeDate).Scan(&route.ID, &route.TripID, &route.Title, &route.RouteDate, &route.SortOrder, &route.CreatedAt)
	return route, err
}

func (s *Store) UpdateRoute(ctx context.Context, tripID, routeID, title string, routeDate *time.Time) (models.TripRoute, error) {
	var route models.TripRoute
	err := s.DB.QueryRow(ctx, `
		UPDATE trip_routes
		SET title = $3,
		    route_date = $4
		WHERE trip_id = $1 AND id = $2
		RETURNING id, trip_id, title, route_date, sort_order, created_at
	`, tripID, routeID, title, routeDate).Scan(&route.ID, &route.TripID, &route.Title, &route.RouteDate, &route.SortOrder, &route.CreatedAt)
	return route, err
}

func (s *Store) DeleteRoute(ctx context.Context, tripID, routeID string) error {
	ct, err := s.DB.Exec(ctx, `DELETE FROM trip_routes WHERE trip_id = $1 AND id = $2`, tripID, routeID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) IsRouteInTrip(ctx context.Context, tripID, routeID string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM trip_routes WHERE trip_id = $1 AND id = $2)
	`, tripID, routeID).Scan(&exists)
	return exists, err
}

func (s *Store) GetRoute(ctx context.Context, tripID, routeID string) (models.TripRoute, error) {
	var route models.TripRoute
	err := s.DB.QueryRow(ctx, `
		SELECT id, trip_id, title, route_date, sort_order, created_at
		FROM trip_routes
		WHERE trip_id = $1 AND id = $2
	`, tripID, routeID).Scan(&route.ID, &route.TripID, &route.Title, &route.RouteDate, &route.SortOrder, &route.CreatedAt)
	return route, err
}

func (s *Store) CountRouteLinks(ctx context.Context, tripID, routeID string) (int, int, error) {
	var locationsCount, expensesCount int
	err := s.DB.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM locations WHERE trip_id = $1 AND route_id = $2),
			(SELECT COUNT(*) FROM expenses WHERE trip_id = $1 AND route_id = $2)
	`, tripID, routeID).Scan(&locationsCount, &expensesCount)
	return locationsCount, expensesCount, err
}
