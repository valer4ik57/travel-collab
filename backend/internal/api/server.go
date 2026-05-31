package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	appauth "travel-collab/backend/internal/auth"
	"travel-collab/backend/internal/config"
	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/repository"
	"travel-collab/backend/internal/ws"
)

type Server struct {
	cfg    config.Config
	store  *repository.Store
	auth   *appauth.Service
	hub    *ws.Hub
	router http.Handler
}

func NewServer(cfg config.Config, store *repository.Store, authService *appauth.Service, hub *ws.Hub) *Server {
	s := &Server{cfg: cfg, store: store, auth: authService, hub: hub}
	s.router = s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RequestSize(s.cfg.MaxRequestBodyBytes))
	r.Use(appmiddleware.CORS(s.cfg))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	authLimiter := appmiddleware.NewRateLimiter(s.cfg.AuthRateLimitRequests, s.cfg.AuthRateLimitWindow)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "travel-collab-backend"})
		})
		r.With(authLimiter.Middleware).Post("/auth/register", s.handleRegister)
		r.With(authLimiter.Middleware).Post("/auth/login", s.handleLogin)
		r.Get("/auth/github", s.handleGitHubStart)
		r.Get("/auth/github/callback", s.handleGitHubCallback)
		r.Get("/ws/{trip_id}", s.handleWebSocket)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Auth(s.auth))
			r.Get("/me", s.handleMe)
			r.Put("/me", s.handleUpdateMe)
			r.Get("/trips", s.handleListTrips)
			r.Post("/trips", s.handleCreateTrip)
			r.Post("/trips/join", s.handleJoinTrip)
			r.Get("/trips/{trip_id}", s.handleGetTrip)
			r.Post("/trips/{trip_id}/leave", s.handleLeaveTrip)
			r.Patch("/trips/{trip_id}/members/{user_id}", s.handleUpdateMemberRole)
			r.Delete("/trips/{trip_id}/members/{user_id}", s.handleRemoveMember)
			r.Get("/trips/{trip_id}/routes", s.handleListRoutes)
			r.Post("/trips/{trip_id}/routes", s.handleCreateRoute)
			r.Put("/trips/{trip_id}/routes/{route_id}", s.handleUpdateRoute)
			r.Delete("/trips/{trip_id}/routes/{route_id}", s.handleDeleteRoute)
			r.Post("/trips/{trip_id}/routes/{route_id}/locations/reorder", s.handleReorderRouteLocations)
			r.Get("/trips/{trip_id}/locations", s.handleListLocations)
			r.Post("/trips/{trip_id}/locations", s.handleCreateLocation)
			r.Put("/trips/{trip_id}/locations/{location_id}", s.handleUpdateLocation)
			r.Delete("/trips/{trip_id}/locations/{location_id}", s.handleDeleteLocation)
			r.Get("/trips/{trip_id}/expenses", s.handleListExpenses)
			r.Post("/trips/{trip_id}/expenses", s.handleCreateExpense)
			r.Put("/trips/{trip_id}/expenses/{expense_id}", s.handleUpdateExpense)
			r.Delete("/trips/{trip_id}/expenses/{expense_id}", s.handleDeleteExpense)
			r.Get("/trips/{trip_id}/messages", s.handleListMessages)
		})
	})

	return r
}
