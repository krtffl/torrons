package http

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// trustedProxyResolver rewrites r.RemoteAddr to the real client IP, honoring
// X-Forwarded-For / X-Real-IP ONLY when the direct TCP peer is one of the
// configured trusted proxies. It replaces chi's middleware.RealIP, which trusts
// those headers from ANY client and so let a client spoof its IP to bypass the
// per-IP rate limiter. Downstream limiters (httprate.KeyByIP) then key off a
// value the client cannot forge.
type trustedProxyResolver struct {
	nets []*net.IPNet
}

// newTrustedProxyResolver parses the trusted-proxy CIDRs; invalid entries are
// logged and skipped rather than failing startup.
func newTrustedProxyResolver(cidrs []string) *trustedProxyResolver {
	r := &trustedProxyResolver{}
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			logger.Warn("[RealIP] Ignoring invalid trusted-proxy CIDR %q: %v", c, err)
			continue
		}
		r.nets = append(r.nets, n)
	}
	return r
}

func (t *trustedProxyResolver) isTrusted(ip net.IP) bool {
	for _, n := range t.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// clientIP resolves the real client IP for the request. If the direct peer is
// not a trusted proxy (or is unparseable), forwarding headers are ignored and
// the real TCP peer address is used — a public client cannot spoof it. If the
// peer IS trusted, the client is taken from the right-most X-Forwarded-For entry
// that is not itself a trusted proxy (proxies append the real peer on the right,
// so this is the client as seen by our outermost trusted proxy and cannot be
// forged by the client's own left-most entries), falling back to X-Real-IP.
func (t *trustedProxyResolver) clientIP(r *http.Request) string {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = h
	}
	peer := net.ParseIP(host)
	if peer == nil || !t.isTrusted(peer) {
		return host
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			p := strings.TrimSpace(parts[i])
			ip := net.ParseIP(p)
			if ip == nil || t.isTrusted(ip) {
				continue
			}
			return p
		}
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" && net.ParseIP(xr) != nil {
		return xr
	}
	return host
}

func (t *trustedProxyResolver) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = t.clientIP(r)
		next.ServeHTTP(w, r)
	})
}

// defaultHTMLContentType sets a default text/html Content-Type for the web-UI
// route group. The template handlers write via buf.WriteTo without setting a
// Content-Type, so Go only sniffs it at write time — too late for the
// compression middleware, which decides whether to compress from the header it
// sees at the first write and so was skipping every HTML page. Setting it up
// front makes those pages compressible. Handlers that emit something else
// override it with their own explicit Set (render.JSON, the PNG endpoints, the
// robots/sitemap/llms handlers). Not applied to /public/* — the file server's
// ServeContent only auto-detects a type when the header is empty, so a preset
// value there would mislabel CSS/images.
func defaultHTMLContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// Context keys for storing user information in request context
type contextKey string

const (
	userIDKey contextKey = "user_id"
)

// lastSeenStaleAfter bounds how often UserMiddleware actually writes a user's
// LastSeen - see isLastSeenStale.
const lastSeenStaleAfter = 5 * time.Minute

// isLastSeenStale reports whether a user's stored LastSeen is old enough to be
// worth refreshing. LastSeen comes back from a Postgres "timestamp without
// time zone" column via database/sql's generic time.Time->string conversion,
// which yields an RFC3339 string (the same conversion behavior documented
// against LastVoteDate in handler.go). If it can't be parsed, default to
// stale so a parse hiccup never silently stops LastSeen from ever refreshing.
func isLastSeenStale(lastSeen string) bool {
	t, err := time.Parse(time.RFC3339, lastSeen)
	if err != nil {
		return true
	}
	return time.Since(t) > lastSeenStaleAfter
}

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

				// Update last seen timestamp asynchronously, but only if it's
				// actually stale. A returning visitor loading several pages in
				// quick succession previously spawned one goroutine + one DB
				// write PER REQUEST for a timestamp that barely moved;
				// isLastSeenStale bounds this to at most one write per user
				// per lastSeenStaleAfter.
				if isLastSeenStale(user.LastSeen) {
					go func() {
						ctx := context.Background()
						if err := h.userRepo.UpdateLastSeen(ctx, userId); err != nil {
							logger.Warn("[UserMiddleware] Failed to update last seen for user %s: %v", userId, err)
						}
					}()
				}
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

// staticPageCacheMaxAge is how long purely-static, zero-DB pages (index, the
// About/FAQ/IGP/comparison/glossary cluster) may be cached. Matches the TTL
// already used for /public/* CSS/JS: short enough that a deploy's content
// change propagates quickly, long enough to actually get cached.
const staticPageCacheMaxAge = "public, max-age=3600"

// setStaticPageCacheHeaders marks a response as cacheable for the handlers
// above. Vary: HX-Request matters here specifically because these templates
// render a full page shell or an htmx partial depending on that header - a
// shared/public cache keys primarily on URL, so without Vary it could serve a
// full-page response to an htmx boost request (or vice versa) cached by a
// different client.
func setStaticPageCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", staticPageCacheMaxAge)
	w.Header().Set("Vary", "HX-Request")
}
