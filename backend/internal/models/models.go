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

type TripRoute struct {
	ID        string     `json:"id"`
	TripID    string     `json:"trip_id"`
	Title     string     `json:"title"`
	RouteDate *time.Time `json:"route_date"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
}

type TripMember struct {
	TripID      string    `json:"trip_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	JoinedAt    time.Time `json:"joined_at"`
	DisplayName string    `json:"display_name"`
	Email       *string   `json:"email"`
	AvatarURL   *string   `json:"avatar_url"`
}

type Location struct {
	ID          string     `json:"id"`
	TripID      string     `json:"trip_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	Category    string     `json:"category"`
	VisitAt     *time.Time `json:"visit_at"`
	RouteID     *string    `json:"route_id"`
	RouteTitle  *string    `json:"route_title,omitempty"`
	RouteDate   *time.Time `json:"route_date,omitempty"`
	RouteOrder  *int       `json:"route_order"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ExpensePayment struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name,omitempty"`
	Amount      float64 `json:"amount"`
}

type ExpenseShare struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name,omitempty"`
	Amount      float64 `json:"amount"`
}

type Expense struct {
	ID           string           `json:"id"`
	TripID       string           `json:"trip_id"`
	Description  string           `json:"description"`
	Amount       float64          `json:"amount"`
	Currency     string           `json:"currency"`
	PaidBy       string           `json:"paid_by,omitempty"`
	PaidByName   string           `json:"paid_by_name,omitempty"`
	SplitWith    []string         `json:"split_with"`
	Payments     []ExpensePayment `json:"payments"`
	Shares       []ExpenseShare   `json:"shares"`
	SplitMode    string           `json:"split_mode"`
	RouteID      *string          `json:"route_id"`
	RouteTitle   *string          `json:"route_title,omitempty"`
	RouteDate    *time.Time       `json:"route_date,omitempty"`
	LocationID   *string          `json:"location_id"`
	LocationName *string          `json:"location_name,omitempty"`
	ExpenseAt    *time.Time       `json:"expense_at"`
	CreatedAt    time.Time        `json:"created_at"`
}

type Balance struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type ExpenseParticipantSummary struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	PaidTotal   float64 `json:"paid_total"`
	ShareTotal  float64 `json:"share_total"`
	NetBalance  float64 `json:"net_balance"`
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

type RouteExpenseSummary struct {
	RouteID       *string    `json:"route_id"`
	RouteTitle    string     `json:"route_title"`
	RouteDate     *time.Time `json:"route_date"`
	TotalAmount   float64    `json:"total_amount"`
	ExpensesCount int        `json:"expenses_count"`
	Currency      string     `json:"currency"`
}

type ExpensesSummary struct {
	Expenses       []Expense                   `json:"expenses"`
	Balances       []Balance                   `json:"balances"`
	Participants   []ExpenseParticipantSummary `json:"participants"`
	Settlements    []Settlement                `json:"settlements"`
	RouteSummaries []RouteExpenseSummary       `json:"route_summaries"`
	TotalAmount    float64                     `json:"total_amount"`
	TotalPaid      float64                     `json:"total_paid"`
	TotalShares    float64                     `json:"total_shares"`
	Currency       string                      `json:"currency"`
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
	Routes    []TripRoute  `json:"routes"`
	Locations []Location   `json:"locations"`
}

type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
