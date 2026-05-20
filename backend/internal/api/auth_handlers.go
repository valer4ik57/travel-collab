package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	appauth "travel-collab/backend/internal/auth"
	appmiddleware "travel-collab/backend/internal/middleware"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Email == "" || !strings.Contains(req.Email, "@") || len(req.Password) < 6 || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "email, password >= 6 and display_name are required")
		return
	}
	hash, err := appauth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user, err := s.store.CreateUser(r.Context(), &req.Email, &hash, req.DisplayName, nil, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "user already exists or data is invalid")
		return
	}
	token, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"token": token, "user": user})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, hash, err := s.store.GetUserByEmailWithPassword(r.Context(), email)
	if err != nil || !appauth.CheckPassword(req.Password, hash) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	token, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"token": token, "user": user})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) githubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.cfg.GitHubClientID,
		ClientSecret: s.cfg.GitHubClientSecret,
		RedirectURL:  s.cfg.GitHubCallbackURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func (s *Server) handleGitHubStart(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusNotImplemented, "GitHub OAuth is not configured")
		return
	}
	state := fmt.Sprintf("travel-collab-%d", time.Now().UnixNano())
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	http.Redirect(w, r, s.githubOAuthConfig().AuthCodeURL(state), http.StatusFound)
}

type githubProfile struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	AvatarURL *string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		writeError(w, http.StatusNotImplemented, "GitHub OAuth is not configured")
		return
	}
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	token, err := s.githubOAuthConfig().Exchange(ctx, code)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to exchange oauth code")
		return
	}
	client := s.githubOAuthConfig().Client(ctx, token)
	profile, err := fetchGitHubProfile(ctx, client)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch github profile")
		return
	}
	email := profile.Email
	if email == nil || *email == "" {
		email = fetchPrimaryGitHubEmail(ctx, client)
	}
	displayName := profile.Login
	if profile.Name != nil && strings.TrimSpace(*profile.Name) != "" {
		displayName = strings.TrimSpace(*profile.Name)
	}
	githubID := fmt.Sprintf("%d", profile.ID)
	user, err := s.store.FindOrCreateGitHubUser(ctx, githubID, displayName, email, profile.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save github user")
		return
	}
	jwtToken, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}
	http.Redirect(w, r, s.cfg.FrontendURL+"/oauth-callback?token="+jwtToken, http.StatusFound)
}

func fetchGitHubProfile(ctx context.Context, client *http.Client) (githubProfile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	res, err := client.Do(req)
	if err != nil {
		return githubProfile{}, err
	}
	defer res.Body.Close()
	var profile githubProfile
	return profile, json.NewDecoder(res.Body).Decode(&profile)
}

func fetchPrimaryGitHubEmail(ctx context.Context, client *http.Client) *string {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	var emails []githubEmail
	if err := json.NewDecoder(res.Body).Decode(&emails); err != nil {
		return nil
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return &e.Email
		}
	}
	return nil
}
