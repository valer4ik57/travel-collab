package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"

	"travel-collab/backend/internal/models"
)

const inviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateInviteCode() (string, error) {
	bytes := make([]byte, 8)
	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteAlphabet))))
		if err != nil {
			return "", err
		}
		bytes[i] = inviteAlphabet[n.Int64()]
	}
	return string(bytes), nil
}

func (s *Store) CreateTrip(ctx context.Context, ownerID, name string, description *string, startDate *time.Time, endDate *time.Time) (models.Trip, error) {
	var trip models.Trip
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return trip, err
	}
	defer tx.Rollback(ctx)

	var inviteCode string
	for attempts := 0; attempts < 5; attempts++ {
		inviteCode, err = generateInviteCode()
		if err != nil {
			return trip, err
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO trips (name, description, owner_id, invite_code, start_date, end_date)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, name, description, owner_id, invite_code, start_date, end_date, created_at
		`, name, description, ownerID, inviteCode, startDate, endDate).Scan(&trip.ID, &trip.Name, &trip.Description, &trip.OwnerID, &trip.InviteCode, &trip.StartDate, &trip.EndDate, &trip.CreatedAt)
		if err == nil {
			break
		}
		if attempts == 4 {
			return trip, err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO trip_members (trip_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, trip.ID, ownerID)
	if err != nil {
		return trip, err
	}
	return trip, tx.Commit(ctx)
}

func (s *Store) ListTripsForUser(ctx context.Context, userID string) ([]models.Trip, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT t.id, t.name, t.description, t.owner_id, t.invite_code, t.start_date, t.end_date, t.created_at,
		       COUNT(DISTINCT tm2.user_id) AS members_count,
		       COUNT(DISTINCT l.id) AS locations_count
		FROM trips t
		JOIN trip_members tm ON tm.trip_id = t.id AND tm.user_id = $1
		LEFT JOIN trip_members tm2 ON tm2.trip_id = t.id
		LEFT JOIN locations l ON l.trip_id = t.id
		GROUP BY t.id
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trips := make([]models.Trip, 0)
	for rows.Next() {
		var trip models.Trip
		if err := rows.Scan(&trip.ID, &trip.Name, &trip.Description, &trip.OwnerID, &trip.InviteCode, &trip.StartDate, &trip.EndDate, &trip.CreatedAt, &trip.MembersCount, &trip.LocationsCount); err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	return trips, rows.Err()
}

func (s *Store) GetTrip(ctx context.Context, tripID string) (models.Trip, error) {
	var trip models.Trip
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, owner_id, invite_code, start_date, end_date, created_at
		FROM trips WHERE id = $1
	`, tripID).Scan(&trip.ID, &trip.Name, &trip.Description, &trip.OwnerID, &trip.InviteCode, &trip.StartDate, &trip.EndDate, &trip.CreatedAt)
	return trip, err
}

func (s *Store) GetTripDetails(ctx context.Context, tripID string) (models.TripDetails, error) {
	trip, err := s.GetTrip(ctx, tripID)
	if err != nil {
		return models.TripDetails{}, err
	}
	members, err := s.ListTripMembers(ctx, tripID)
	if err != nil {
		return models.TripDetails{}, err
	}
	locations, err := s.ListLocations(ctx, tripID)
	if err != nil {
		return models.TripDetails{}, err
	}
	return models.TripDetails{Trip: trip, Members: members, Locations: locations}, nil
}

func (s *Store) IsTripMember(ctx context.Context, tripID, userID string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM trip_members WHERE trip_id = $1 AND user_id = $2)
	`, tripID, userID).Scan(&exists)
	return exists, err
}

func (s *Store) ListTripMembers(ctx context.Context, tripID string) ([]models.TripMember, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT tm.trip_id, tm.user_id, tm.role, tm.joined_at, u.display_name, u.email, u.avatar_url
		FROM trip_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.trip_id = $1
		ORDER BY CASE tm.role WHEN 'owner' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END, tm.joined_at ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]models.TripMember, 0)
	for rows.Next() {
		var m models.TripMember
		if err := rows.Scan(&m.TripID, &m.UserID, &m.Role, &m.JoinedAt, &m.DisplayName, &m.Email, &m.AvatarURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) JoinTripByInviteCode(ctx context.Context, userID, inviteCode string) (models.Trip, models.TripMember, bool, error) {
	var trip models.Trip
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return trip, models.TripMember{}, false, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT id, name, description, owner_id, invite_code, start_date, end_date, created_at
		FROM trips WHERE invite_code = $1
	`, inviteCode).Scan(&trip.ID, &trip.Name, &trip.Description, &trip.OwnerID, &trip.InviteCode, &trip.StartDate, &trip.EndDate, &trip.CreatedAt)
	if err != nil {
		return trip, models.TripMember{}, false, err
	}

	ct, err := tx.Exec(ctx, `
		INSERT INTO trip_members (trip_id, user_id, role)
		VALUES ($1, $2, 'editor')
		ON CONFLICT (trip_id, user_id) DO NOTHING
	`, trip.ID, userID)
	if err != nil {
		return trip, models.TripMember{}, false, err
	}
	created := ct.RowsAffected() > 0

	var member models.TripMember
	err = tx.QueryRow(ctx, `
		SELECT tm.trip_id, tm.user_id, tm.role, tm.joined_at, u.display_name, u.email, u.avatar_url
		FROM trip_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.trip_id = $1 AND tm.user_id = $2
	`, trip.ID, userID).Scan(&member.TripID, &member.UserID, &member.Role, &member.JoinedAt, &member.DisplayName, &member.Email, &member.AvatarURL)
	if err != nil {
		return trip, models.TripMember{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return trip, models.TripMember{}, false, err
	}
	return trip, member, created, nil
}

func (s *Store) EnsureAllUsersAreTripMembers(ctx context.Context, tripID string, userIDs []string) error {
	for _, uid := range userIDs {
		ok, err := s.IsTripMember(ctx, tripID, uid)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("user %s is not a trip member", uid)
		}
	}
	return nil
}
