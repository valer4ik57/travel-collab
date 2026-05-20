package models

import "time"

type User struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	GitHubID    *string   `json:"github_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Trip struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	OwnerID        string     `json:"owner_id"`
	InviteCode     string     `json:"invite_code"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	CreatedAt      time.Time  `json:"created_at"`
	MembersCount   int        `json:"members_count,omitempty"`
	LocationsCount int        `json:"locations_count,omitempty"`
}

type TripMember struct {
	TripID      string    `json:"trip_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	DisplayName string    `json:"display_name"`
	Email       *string   `json:"email"`
	AvatarURL   *string   `json:"avatar_url"`
}

type Location struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Category    string    `json:"category"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Expense struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaidBy      string    `json:"paid_by"`
	PaidByName  string    `json:"paid_by_name,omitempty"`
	SplitWith   []string  `json:"split_with"`
	CreatedAt   time.Time `json:"created_at"`
}

type Balance struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type Settlement struct {
	FromUserID string  `json:"from_user_id"`
	FromName   string  `json:"from_name"`
	ToUserID   string  `json:"to_user_id"`
	ToName     string  `json:"to_name"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type ExpensesSummary struct {
	Expenses    []Expense    `json:"expenses"`
	Balances    []Balance    `json:"balances"`
	Settlements []Settlement `json:"settlements"`
}

type Message struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Text        string    `json:"text"`
	SentAt      time.Time `json:"sent_at"`
}

type TripDetails struct {
	Trip      Trip         `json:"trip"`
	Members   []TripMember `json:"members"`
	Locations []Location   `json:"locations"`
}

type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
