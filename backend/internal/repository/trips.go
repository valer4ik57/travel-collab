package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"

	"travel-collab/backend/internal/models"
)

var ErrTripMemberRemoved = errors.New("trip member was removed")

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
		INSERT INTO trip_members (trip_id, user_id, role, status)
		VALUES ($1, $2, 'owner', 'active')
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
		JOIN trip_members tm ON tm.trip_id = t.id AND tm.user_id = $1 AND tm.status = 'active'
		LEFT JOIN trip_members tm2 ON tm2.trip_id = t.id AND tm2.status = 'active'
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
		SELECT EXISTS(SELECT 1 FROM trip_members WHERE trip_id = $1 AND user_id = $2 AND status = 'active')
	`, tripID, userID).Scan(&exists)
	return exists, err
}

func (s *Store) GetTripMemberRole(ctx context.Context, tripID, userID string) (string, error) {
	var role string
	err := s.DB.QueryRow(ctx, `
		SELECT role FROM trip_members WHERE trip_id = $1 AND user_id = $2 AND status = 'active'
	`, tripID, userID).Scan(&role)
	return role, err
}

func (s *Store) ListTripMembers(ctx context.Context, tripID string) ([]models.TripMember, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT tm.trip_id, tm.user_id, tm.role, tm.status, tm.joined_at, u.display_name, u.email, u.avatar_url
		FROM trip_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.trip_id = $1 AND tm.status = 'active'
		ORDER BY CASE tm.role WHEN 'owner' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END, tm.joined_at ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]models.TripMember, 0)
	for rows.Next() {
		var m models.TripMember
		if err := rows.Scan(&m.TripID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.DisplayName, &m.Email, &m.AvatarURL); err != nil {
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

	var oldRole, oldStatus string
	err = tx.QueryRow(ctx, `
		SELECT role, status FROM trip_members WHERE trip_id = $1 AND user_id = $2
	`, trip.ID, userID).Scan(&oldRole, &oldStatus)

	created := false
	if err == pgx.ErrNoRows {
		_, err = tx.Exec(ctx, `
			INSERT INTO trip_members (trip_id, user_id, role, status, joined_at)
			VALUES ($1, $2, 'editor', 'active', now())
		`, trip.ID, userID)
		created = true
	} else if err == nil {
		if oldStatus == "removed" {
			return trip, models.TripMember{}, false, ErrTripMemberRemoved
		}
		if oldStatus == "left" {
			_, err = tx.Exec(ctx, `
				UPDATE trip_members SET status = 'active', joined_at = now()
				WHERE trip_id = $1 AND user_id = $2
			`, trip.ID, userID)
			created = true
		}
	} else {
		return trip, models.TripMember{}, false, err
	}
	if err != nil {
		return trip, models.TripMember{}, false, err
	}

	member, err := scanTripMember(ctx, tx, trip.ID, userID)
	if err != nil {
		return trip, models.TripMember{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return trip, models.TripMember{}, false, err
	}
	return trip, member, created, nil
}

func scanTripMember(ctx context.Context, tx pgx.Tx, tripID, userID string) (models.TripMember, error) {
	var member models.TripMember
	err := tx.QueryRow(ctx, `
		SELECT tm.trip_id, tm.user_id, tm.role, tm.status, tm.joined_at, u.display_name, u.email, u.avatar_url
		FROM trip_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.trip_id = $1 AND tm.user_id = $2
	`, tripID, userID).Scan(&member.TripID, &member.UserID, &member.Role, &member.Status, &member.JoinedAt, &member.DisplayName, &member.Email, &member.AvatarURL)
	return member, err
}

func (s *Store) EnsureAllUsersAreTripMembers(ctx context.Context, tripID string, userIDs []string) error {
	for _, uid := range userIDs {
		ok, err := s.IsTripMember(ctx, tripID, uid)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("user %s is not an active trip member", uid)
		}
	}
	return nil
}

func (s *Store) UpdateTripMemberRole(ctx context.Context, tripID, targetUserID, role string) (models.TripMember, error) {
	var member models.TripMember
	err := s.DB.QueryRow(ctx, `
		UPDATE trip_members
		SET role = $3
		WHERE trip_id = $1 AND user_id = $2 AND status = 'active'
		RETURNING trip_id, user_id, role, status, joined_at
	`, tripID, targetUserID, role).Scan(&member.TripID, &member.UserID, &member.Role, &member.Status, &member.JoinedAt)
	if err != nil {
		return member, err
	}
	user, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return member, err
	}
	member.DisplayName = user.DisplayName
	member.Email = user.Email
	member.AvatarURL = user.AvatarURL
	return member, nil
}

func (s *Store) LeaveTripMember(ctx context.Context, tripID, targetUserID string) error {
	ct, err := s.DB.Exec(ctx, `
		UPDATE trip_members SET status = 'left'
		WHERE trip_id = $1 AND user_id = $2 AND status = 'active'
	`, tripID, targetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) RemoveTripMember(ctx context.Context, tripID, targetUserID string) error {
	ct, err := s.DB.Exec(ctx, `
		UPDATE trip_members SET status = 'removed'
		WHERE trip_id = $1 AND user_id = $2 AND status = 'active'
	`, tripID, targetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) CountOwners(ctx context.Context, tripID string) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM trip_members WHERE trip_id = $1 AND role = 'owner' AND status = 'active'
	`, tripID).Scan(&count)
	return count, err
}
