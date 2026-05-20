package repository

import (
	"context"

	"travel-collab/backend/internal/models"
)

func (s *Store) CreateUser(ctx context.Context, email *string, passwordHash *string, displayName string, avatarURL *string, githubID *string) (models.User, error) {
	var user models.User
	err := s.DB.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url, github_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, display_name, avatar_url, github_id, created_at
	`, email, passwordHash, displayName, avatarURL, githubID).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.GitHubID, &user.CreatedAt)
	return user, err
}

func (s *Store) GetUserByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	err := s.DB.QueryRow(ctx, `
		SELECT id, email, display_name, avatar_url, github_id, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.GitHubID, &user.CreatedAt)
	return user, err
}

func (s *Store) GetUserByEmailWithPassword(ctx context.Context, email string) (models.User, string, error) {
	var user models.User
	var passwordHash string
	err := s.DB.QueryRow(ctx, `
		SELECT id, email, display_name, avatar_url, github_id, created_at, password_hash
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.GitHubID, &user.CreatedAt, &passwordHash)
	return user, passwordHash, err
}

func (s *Store) FindOrCreateGitHubUser(ctx context.Context, githubID, displayName string, email *string, avatarURL *string) (models.User, error) {
	var user models.User
	err := s.DB.QueryRow(ctx, `
		INSERT INTO users (github_id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (github_id) DO UPDATE SET
			email = COALESCE(users.email, EXCLUDED.email),
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url
		RETURNING id, email, display_name, avatar_url, github_id, created_at
	`, githubID, email, displayName, avatarURL).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.GitHubID, &user.CreatedAt)
	return user, err
}
