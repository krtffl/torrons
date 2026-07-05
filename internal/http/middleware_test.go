package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRequireAdminToken covers Handler.RequireAdminToken's fail-closed
// behavior: it must reject every request when no ADMIN_TOKEN is configured,
// and otherwise only accept a well-formed "Authorization: Bearer <token>"
// header matching the configured token exactly.
func TestRequireAdminToken(t *testing.T) {
	tests := []struct {
		name           string
		adminToken     string
		authHeader     string
		setAuthHeader  bool
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "not configured - well-formed header still rejected",
			adminToken:     "",
			authHeader:     "Bearer anything",
			setAuthHeader:  true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "configured - no Authorization header at all",
			adminToken:     "test-secret-123",
			setAuthHeader:  false,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "configured - header missing Bearer prefix (raw token)",
			adminToken:     "test-secret-123",
			authHeader:     "test-secret-123",
			setAuthHeader:  true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "configured - header missing Bearer prefix (Basic scheme)",
			adminToken:     "test-secret-123",
			authHeader:     "Basic dGVzdC1zZWNyZXQtMTIz",
			setAuthHeader:  true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "configured - wrong token",
			adminToken:     "test-secret-123",
			authHeader:     "Bearer wrong-token",
			setAuthHeader:  true,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "configured - correct token",
			adminToken:     "test-secret-123",
			authHeader:     "Bearer test-secret-123",
			setAuthHeader:  true,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{adminToken: tt.adminToken}

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/bracket/1/create?size=2", nil)
			if tt.setAuthHeader {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			h.RequireAdminToken(next).ServeHTTP(rec, req)

			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
