package http

import (
	"context"
	"net/http"

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
			HttpOnly: true,                    // Prevent JavaScript access (XSS protection)
			Secure:   r.TLS != nil,            // Only send over HTTPS in production
			SameSite: http.SameSiteLaxMode,   // CSRF protection
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
