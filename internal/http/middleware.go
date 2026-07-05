package http

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// Context keys for storing user information in request context
type contextKey string

const (
	userIDKey contextKey = "user_id"
)

// UserMiddleware handles user identification via cookies
// Creates new users if cookie doesn't exist, validates existing users
func (h *Handler) UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The embeddable leaderboard widget is read-only and rendered
		// cross-origin inside third-party <iframe>s, where third-party
		// cookies are commonly blocked by the browser. Without this early
		// return, every cookie-less impression would mint a brand new
		// anonymous Users row on every page load - the widget doesn't need
		// a user identity at all (it only reads torroRepo), so skip
		// cookie/user-creation entirely for this path.
		if strings.HasPrefix(r.URL.Path, "/embed/") {
			next.ServeHTTP(w, r)
			return
		}

		const cookieName = "torrons_user_id"
		const cookieMaxAge = 90 * 24 * 60 * 60 // 90 days in seconds

		var userId string

		// Try to get existing user ID from cookie
		cookie, err := r.Cookie(cookieName)
		if err == nil && cookie.Value != "" {
			// Validate that user exists in database
			user, err := h.userRepo.Get(r.Context(), cookie.Value)
			if err == nil && user != nil {
				userId = user.Id

				// Update last seen timestamp (async, don't block request)
				go func() {
					ctx := context.Background()
					if err := h.userRepo.UpdateLastSeen(ctx, userId); err != nil {
						logger.Warn("[UserMiddleware] Failed to update last seen for user %s: %v", userId, err)
					}
				}()
			} else {
				// Cookie exists but user not found in DB, create new user
				logger.Info("[UserMiddleware] Cookie found but user not in DB, creating new user")
				userId = ""
			}
		}

		// Create new user if no valid userId found
		if userId == "" {
			userId = uuid.NewString()

			newUser := &domain.User{
				Id:        userId,
				VoteCount: 0,
			}

			createdUser, err := h.userRepo.Create(r.Context(), newUser)
			if err != nil {
				logger.Error("[UserMiddleware] Failed to create user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			userId = createdUser.Id
			logger.Info("[UserMiddleware] Created new user: %s", userId)
		}

		// Set cookie (refresh expiration even for existing users)
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    userId,
			Path:     "/",
			MaxAge:   cookieMaxAge,
			HttpOnly: true,                 // Prevent JavaScript access (XSS protection)
			Secure:   r.TLS != nil,         // Only send over HTTPS in production
			SameSite: http.SameSiteLaxMode, // CSRF protection
		})

		// Add user ID to request context for handlers to use
		ctx := context.WithValue(r.Context(), userIDKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the user ID from request context
// Returns empty string if not found
func GetUserIDFromContext(ctx context.Context) string {
	userId, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return userId
}

// RequireAdminToken gates admin-only routes behind a shared-secret bearer
// token configured via ADMIN_TOKEN (see internal/config). There is no
// broader user/role system in this codebase - see UserMiddleware above -
// this is a deliberately minimal single-shared-secret gate for a couple
// of operator-only endpoints, not a general auth system.
//
// Fails closed: if no token is configured, every request is rejected
// rather than silently allowed through. The response never distinguishes
// "not configured" from "wrong token" to the caller - only server logs
// differentiate the two, so a caller can't use the error to fingerprint
// server configuration state.
func (h *Handler) RequireAdminToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		authHeader := r.Header.Get("Authorization")
		token, hasPrefix := strings.CutPrefix(authHeader, prefix)

		valid := h.adminToken != "" && hasPrefix &&
			subtle.ConstantTimeCompare([]byte(token), []byte(h.adminToken)) == 1

		if !valid {
			if h.adminToken == "" {
				logger.Error("[RequireAdminToken] ADMIN_TOKEN is not configured; rejecting admin request to %s", r.URL.Path)
			} else {
				logger.Warn("[RequireAdminToken] Rejected admin request to %s from %s", r.URL.Path, r.RemoteAddr)
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			render.Render(w, r, domain.ErrUnauthorized(fmt.Errorf("%s: invalid or missing admin token", domain.UnauthorizedError)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
