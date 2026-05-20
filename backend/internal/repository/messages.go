package repository

import (
	"context"

	"travel-collab/backend/internal/models"
)

func (s *Store) ListMessages(ctx context.Context, tripID string, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.DB.Query(ctx, `
		SELECT * FROM (
			SELECT m.id, m.trip_id, m.user_id, u.display_name, u.avatar_url, m.text, m.sent_at
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.trip_id = $1
			ORDER BY m.sent_at DESC
			LIMIT $2
		) x ORDER BY sent_at ASC
	`, tripID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]models.Message, 0)
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.TripID, &m.UserID, &m.DisplayName, &m.AvatarURL, &m.Text, &m.SentAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *Store) SaveMessage(ctx context.Context, tripID, userID, text string) (models.Message, error) {
	var m models.Message
	err := s.DB.QueryRow(ctx, `
		INSERT INTO messages (trip_id, user_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, trip_id, user_id,
		          (SELECT display_name FROM users WHERE id = $2),
		          (SELECT avatar_url FROM users WHERE id = $2),
		          text, sent_at
	`, tripID, userID, text).Scan(&m.ID, &m.TripID, &m.UserID, &m.DisplayName, &m.AvatarURL, &m.Text, &m.SentAt)
	return m, err
}
