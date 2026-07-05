package http

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
)

// -- fakes --
//
// friendsIndex/friendsCreate/friendsJoin/friendsLeaderboard never call
// h.db.Begin() -- they only exercise plain (non-Tx) domain.FriendCircleRepo
// and domain.ClassRepo methods -- so hand-rolled in-memory fakes are enough
// to exercise them without a real database.

// fakeFriendCircleRepo is a minimal, map-backed stand-in for
// domain.FriendCircleRepo. It's only meant to satisfy this test file's
// needs, not to be a reusable test double.
type fakeFriendCircleRepo struct {
	circles []*domain.FriendCircle          // creation order, for deterministic ListForUser
	byId    map[string]*domain.FriendCircle // Id -> circle
	byCode  map[string]string               // InviteCode -> Id
	members map[string]map[string]bool      // circleId -> set(userId)

	// leaderboard is returned as-is by both GetCircleLeaderboard and
	// GetCircleGlobalLeaderboard, configurable per test case.
	leaderboard []*domain.UserLeaderboardEntry

	nextId int
}

func newFakeFriendCircleRepo() *fakeFriendCircleRepo {
	return &fakeFriendCircleRepo{
		byId:    make(map[string]*domain.FriendCircle),
		byCode:  make(map[string]string),
		members: make(map[string]map[string]bool),
	}
}

func (f *fakeFriendCircleRepo) Create(ctx context.Context, ownerUserId string) (*domain.FriendCircle, error) {
	f.nextId++
	circle := &domain.FriendCircle{
		Id:          fmt.Sprintf("circle-%d", f.nextId),
		OwnerUserId: ownerUserId,
		InviteCode:  fmt.Sprintf("invite-code-%d", f.nextId),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	f.circles = append(f.circles, circle)
	f.byId[circle.Id] = circle
	f.byCode[circle.InviteCode] = circle.Id
	f.members[circle.Id] = map[string]bool{ownerUserId: true}

	return circle, nil
}

func (f *fakeFriendCircleRepo) Get(ctx context.Context, id string) (*domain.FriendCircle, error) {
	circle, ok := f.byId[id]
	if !ok {
		return nil, fmt.Errorf("%s: circle %s not found", domain.NotFoundError, id)
	}
	return circle, nil
}

func (f *fakeFriendCircleRepo) GetByInviteCode(ctx context.Context, inviteCode string) (*domain.FriendCircle, error) {
	id, ok := f.byCode[inviteCode]
	if !ok {
		return nil, fmt.Errorf("%s: invite code %s not found", domain.NotFoundError, inviteCode)
	}
	return f.byId[id], nil
}

func (f *fakeFriendCircleRepo) AddMember(ctx context.Context, circleId string, userId string) error {
	if _, ok := f.byId[circleId]; !ok {
		return fmt.Errorf("%s: circle %s not found", domain.NotFoundError, circleId)
	}
	if f.members[circleId] == nil {
		f.members[circleId] = make(map[string]bool)
	}
	// Map-set membership: adding the same userId twice is a no-op, which is
	// exactly the idempotency AddMember's doc comment promises.
	f.members[circleId][userId] = true
	return nil
}

func (f *fakeFriendCircleRepo) IsMember(ctx context.Context, circleId string, userId string) (bool, error) {
	return f.members[circleId][userId], nil
}

func (f *fakeFriendCircleRepo) ListForUser(ctx context.Context, userId string) ([]*domain.FriendCircle, error) {
	var out []*domain.FriendCircle
	for _, circle := range f.circles {
		if f.members[circle.Id][userId] {
			out = append(out, circle)
		}
	}
	return out, nil
}

func (f *fakeFriendCircleRepo) GetCircleLeaderboard(ctx context.Context, circleId string, classId string) ([]*domain.UserLeaderboardEntry, error) {
	return f.leaderboard, nil
}

func (f *fakeFriendCircleRepo) GetCircleGlobalLeaderboard(ctx context.Context, circleId string) ([]*domain.UserLeaderboardEntry, error) {
	return f.leaderboard, nil
}

// memberCount is a test-only helper to inspect membership without going
// through the repo interface.
func (f *fakeFriendCircleRepo) memberCount(circleId string) int {
	return len(f.members[circleId])
}

// fakeUserRepo is a minimal, map-backed stand-in for domain.UserRepo.
// friends_handler.go doesn't actually call into userRepo today (it reads
// the current user solely from context, via GetUserIDFromContext), but the
// Handler struct still needs a non-nil value satisfying the interface.
type fakeUserRepo struct {
	users map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User)}
}

func (f *fakeUserRepo) Get(ctx context.Context, id string) (*domain.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, fmt.Errorf("%s: user %s not found", domain.NotFoundError, id)
	}
	return u, nil
}

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	f.users[user.Id] = user
	return user, nil
}

func (f *fakeUserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	f.users[user.Id] = user
	return user, nil
}

func (f *fakeUserRepo) IncrementVoteCount(ctx context.Context, userId string, classId string) error {
	return nil
}

func (f *fakeUserRepo) GetVoteCountForClass(ctx context.Context, userId string, classId string) (int, error) {
	return 0, nil
}

func (f *fakeUserRepo) UpdateLastSeen(ctx context.Context, userId string) error {
	return nil
}

func (f *fakeUserRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.User, error) {
	return f.Get(ctx, id)
}

func (f *fakeUserRepo) IncrementVoteCountTx(tx *sql.Tx, ctx context.Context, userId string, classId string) error {
	return nil
}

func (f *fakeUserRepo) UpdateStreakTx(tx *sql.Tx, ctx context.Context, userId string) error {
	return nil
}

// fakeClassRepo is a minimal stand-in for domain.ClassRepo, used only by
// friendsLeaderboard's category switcher.
type fakeClassRepo struct {
	classes []*domain.Class
}

func (f *fakeClassRepo) List(ctx context.Context) ([]*domain.Class, error) {
	return f.classes, nil
}

// -- test setup helpers --

// newFriendsTestHandler builds a Handler with real (embedded) templates but
// hand-rolled fake repos, skipping NewHandler entirely since it requires a
// *sql.DB this package's friends_handler.go never actually needs.
func newFriendsTestHandler(t *testing.T, circleRepo *fakeFriendCircleRepo, userRepo *fakeUserRepo, classRepo *fakeClassRepo) *Handler {
	t.Helper()

	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	return &Handler{
		template:         tmpls,
		bpool:            bpool.NewBufferPool(8),
		friendCircleRepo: circleRepo,
		userRepo:         userRepo,
		classRepo:        classRepo,
	}
}

// newFriendsRequest builds an httptest.Request carrying the given URL
// params (as chi would after routing) and an optional user ID in context
// (as UserMiddleware would set). It always sets HX-Request: true so the
// handler renders just the "friends" fragment, not the full HTML page
// (header/topbar aren't relevant to what these tests assert on).
func newFriendsRequest(method, target string, urlParams map[string]string, userId string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("HX-Request", "true")

	rctx := chi.NewRouteContext()
	for k, v := range urlParams {
		rctx.URLParams.Add(k, v)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

	if userId != "" {
		ctx = context.WithValue(ctx, userIDKey, userId)
	}

	return req.WithContext(ctx)
}

// -- friendsCreate --

func TestFriendsCreate(t *testing.T) {
	t.Run("no user in context is rejected", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodPost, "/friends/create", nil, "")
		rec := httptest.NewRecorder()

		h.friendsCreate(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("creates a circle and returns its invite code", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodPost, "/friends/create", nil, "owner-1")
		rec := httptest.NewRecorder()

		h.friendsCreate(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		if len(circleRepo.circles) != 1 {
			t.Fatalf("expected 1 circle to be created, got %d", len(circleRepo.circles))
		}
		circle := circleRepo.circles[0]

		if circle.OwnerUserId != "owner-1" {
			t.Errorf("circle owner = %q, want %q", circle.OwnerUserId, "owner-1")
		}
		if circle.InviteCode == "" {
			t.Error("expected a non-empty invite code to be generated")
		}
		if !circleRepo.members[circle.Id]["owner-1"] {
			t.Error("expected the owner to be added as the circle's first member")
		}

		body := rec.Body.String()
		if !strings.Contains(body, circle.InviteCode) {
			t.Errorf("response body doesn't contain the invite code %q: %s", circle.InviteCode, body)
		}
		if !strings.Contains(body, "/friends/join/"+circle.InviteCode) {
			t.Errorf("response body doesn't contain the invite URL path: %s", body)
		}
	})
}

// -- friendsJoin --

func TestFriendsJoin(t *testing.T) {
	t.Run("no user in context is rejected", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/join/whatever", map[string]string{"inviteCode": "whatever"}, "")
		rec := httptest.NewRecorder()

		h.friendsJoin(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("valid invite code adds the joining user as a member and redirects", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "owner-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/join/"+circle.InviteCode,
			map[string]string{"inviteCode": circle.InviteCode}, "joiner-1")
		rec := httptest.NewRecorder()

		h.friendsJoin(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusFound, rec.Body.String())
		}
		wantLocation := "/friends/" + circle.Id
		if got := rec.Header().Get("Location"); got != wantLocation {
			t.Errorf("Location = %q, want %q", got, wantLocation)
		}

		isMember, err := circleRepo.IsMember(context.Background(), circle.Id, "joiner-1")
		if err != nil {
			t.Fatalf("IsMember: %v", err)
		}
		if !isMember {
			t.Error("expected joiner-1 to be a member of the circle after joining")
		}
	})

	t.Run("joining twice is idempotent", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "owner-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})

		for i := 0; i < 2; i++ {
			req := newFriendsRequest(http.MethodGet, "/friends/join/"+circle.InviteCode,
				map[string]string{"inviteCode": circle.InviteCode}, "joiner-1")
			rec := httptest.NewRecorder()

			h.friendsJoin(rec, req)

			if rec.Code != http.StatusFound {
				t.Fatalf("attempt %d: status = %d, want %d; body: %s", i+1, rec.Code, http.StatusFound, rec.Body.String())
			}
		}

		// Owner + joiner-1 == 2 members; joining a second time must not
		// duplicate the membership or error.
		if got := circleRepo.memberCount(circle.Id); got != 2 {
			t.Errorf("member count after joining twice = %d, want 2", got)
		}
	})

	t.Run("unknown invite code renders the invalid-invite state, not a crash", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/join/does-not-exist",
			map[string]string{"inviteCode": "does-not-exist"}, "joiner-1")
		rec := httptest.NewRecorder()

		h.friendsJoin(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "no és vàlid") {
			t.Errorf("expected the invalid-invite fragment to be rendered, got: %s", rec.Body.String())
		}

		// No circle exists at all, so there's nothing to have joined -- just
		// double-check the repo wasn't mutated by the failed lookup.
		if len(circleRepo.circles) != 0 {
			t.Errorf("expected no circles to exist, got %d", len(circleRepo.circles))
		}
	})
}

// -- friendsIndex --

func TestFriendsIndex(t *testing.T) {
	t.Run("no user in context is rejected", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends", nil, "")
		rec := httptest.NewRecorder()

		h.friendsIndex(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("user with no circles yet sees the empty state, not a crash", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends", nil, "lonely-user")
		rec := httptest.NewRecorder()

		h.friendsIndex(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "Encara no formes part de cap cercle") {
			t.Errorf("expected the no-circles empty state, got: %s", rec.Body.String())
		}
	})

	t.Run("user with circles sees them listed", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends", nil, "user-1")
		rec := httptest.NewRecorder()

		h.friendsIndex(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "/friends/"+circle.Id) {
			t.Errorf("expected the owned circle to be listed, got: %s", rec.Body.String())
		}
	})
}

// -- friendsLeaderboard --

func TestFriendsLeaderboard(t *testing.T) {
	t.Run("no user in context is rejected", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/whatever", map[string]string{"circleId": "whatever"}, "")
		rec := httptest.NewRecorder()

		h.friendsLeaderboard(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unknown circle is a 404", func(t *testing.T) {
		h := newFriendsTestHandler(t, newFakeFriendCircleRepo(), newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/does-not-exist", map[string]string{"circleId": "does-not-exist"}, "user-1")
		rec := httptest.NewRecorder()

		h.friendsLeaderboard(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("non-member sees the not-member state, not a crash", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "owner-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/"+circle.Id, map[string]string{"circleId": circle.Id}, "outsider")
		rec := httptest.NewRecorder()

		h.friendsLeaderboard(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "No formes part d'aquest cercle") {
			t.Errorf("expected the not-member fragment, got: %s", rec.Body.String())
		}
	})

	t.Run("member with no leaderboard data yet sees the empty state, not a crash", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "owner-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}
		circleRepo.leaderboard = nil // no member has voted enough yet

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/"+circle.Id, map[string]string{"circleId": circle.Id}, "owner-1")
		rec := httptest.NewRecorder()

		h.friendsLeaderboard(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "Encara no hi ha prou vots") {
			t.Errorf("expected the empty-leaderboard message, got: %s", rec.Body.String())
		}
	})

	t.Run("member sees the circle's leaderboard entries", func(t *testing.T) {
		circleRepo := newFakeFriendCircleRepo()
		circle, err := circleRepo.Create(context.Background(), "owner-1")
		if err != nil {
			t.Fatalf("setup: failed to create circle: %v", err)
		}
		circleRepo.leaderboard = []*domain.UserLeaderboardEntry{
			{Rank: 1, TorronId: "t1", TorronName: "Torró de Xocolata", Rating: 1600, VoteCount: 5},
		}

		h := newFriendsTestHandler(t, circleRepo, newFakeUserRepo(), &fakeClassRepo{})
		req := newFriendsRequest(http.MethodGet, "/friends/"+circle.Id, map[string]string{"circleId": circle.Id}, "owner-1")
		rec := httptest.NewRecorder()

		h.friendsLeaderboard(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "Torró de Xocolata") {
			t.Errorf("expected the leaderboard entry to be rendered, got: %s", rec.Body.String())
		}
	})
}
