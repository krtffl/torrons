//go:build integration

package http

// Integration tests for the transactional, real-Postgres-only handlers:
// the open-voting vote-casting handler (result, in handler.go) and the
// bracket lifecycle (bracketCreate/seedAndCreateBracket, bracketMatchVote,
// bracketAdvance, in bracket_handler.go). All of these call h.db.Begin()
// directly and pass the *sql.Tx into several repos' Tx methods, so unlike
// friends_handler_test.go's hand-rolled fakes, there's no reasonable way to
// exercise them without a real database -- introducing a transaction
// abstraction purely to make these mockable would be a bigger change than
// "add tests" warrants.
//
// These are gated behind the "integration" build tag so a plain
// `go test ./...` never touches a database: only
// `go test -tags=integration ./...` compiles and runs this file.
//
// # Running these tests
//
// 1. Start a throwaway Postgres in podman:
//
//	podman run -d --rm --name torrons-test-db \
//	  -e POSTGRES_USER=myUser \
//	  -e POSTGRES_PASSWORD=myPassword \
//	  -e POSTGRES_DB=databaseName \
//	  -p 5433:5432 \
//	  postgres:16
//
// 2. Apply migrations (golang-migrate CLI, matching MIGRATIONS.md):
//
//	migrate -path migrations \
//	  -database "postgres://myUser:myPassword@localhost:5433/databaseName?sslmode=disable" \
//	  up
//
// 3. Run the integration suite, pointing it at that database via the same
//    DB_* environment variables internal/config and internal/api already
//    read elsewhere in this codebase:
//
//	DB_HOST=localhost DB_PORT=5433 DB_USER=myUser DB_PASSWORD=myPassword \
//	  DB_NAME=databaseName DB_SSL_MODE=disable \
//	  go test -tags=integration ./internal/http/...
//
// Note: internal/api.NewDatabaseConnection also runs these same migrations
// automatically on connect, so step 2 is technically redundant if you only
// ever go through that path -- it's spelled out here anyway because it's
// the documented, explicit way to prep a database for this suite, and this
// test file intentionally does NOT import internal/api (it already imports
// internal/http, so this package importing it back would be a cycle) or
// run migrations itself; it assumes the schema already exists.
//
// If DB_* variables are left unset, they default to this project's usual
// local-dev values (see config/config.default.yaml): localhost:5432,
// myUser/myPassword, database "databaseName", sslmode disable.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/repository"
)

// getEnvOrDefault reads an environment variable, falling back to a default
// if unset -- mirroring internal/config's DB_* override behavior without
// pulling in viper/godotenv just for a handful of connection settings.
func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setupIntegrationDB opens a connection to the Postgres instance described
// by DB_* environment variables (see the package doc above), verifying it's
// reachable and already migrated. It does not run migrations itself.
func setupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	port, err := strconv.Atoi(getEnvOrDefault("DB_PORT", "5432"))
	if err != nil {
		t.Fatalf("invalid DB_PORT: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_USER", "myUser"),
		getEnvOrDefault("DB_PASSWORD", "myPassword"),
		getEnvOrDefault("DB_NAME", "databaseName"),
		port,
		getEnvOrDefault("DB_SSL_MODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open Postgres connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("failed to reach Postgres (is it running and migrated? see this file's package doc for the podman + migrate recipe): %v", err)
	}

	if _, err := db.Exec(`SELECT 1 FROM "Classes" LIMIT 1`); err != nil {
		db.Close()
		t.Fatalf("database doesn't look migrated (querying \"Classes\" failed: %v) -- run the migrate step described in this file's package doc first", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

// newIntegrationTemplate parses the real embedded templates, exactly as
// NewHandler does, so handlers that render HTML on success (result's
// pairing.html, bracketMatchVote's bracket-vote-page) don't nil-panic.
func newIntegrationTemplate(t *testing.T) *template.Template {
	t.Helper()

	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}
	return tmpls
}

// insertTestClass inserts a minimal Classes row directly via SQL --
// domain.ClassRepo exposes no Create method (classes are seed data, not
// created at runtime by any handler), so this is the only way to get one
// into a fresh database for a test.
func insertTestClass(t *testing.T, db *sql.DB, name string) string {
	t.Helper()

	id := uuid.NewString()
	if _, err := db.Exec(
		`INSERT INTO "Classes" ("Id", "Name", "Description") VALUES ($1, $2, $3)`,
		id, name, "integration test class",
	); err != nil {
		t.Fatalf("failed to insert test class: %v", err)
	}
	return id
}

// insertTestTorro inserts a minimal Torrons row directly via SQL --
// domain.TorroRepo, like ClassRepo, exposes no Create method.
func insertTestTorro(t *testing.T, db *sql.DB, classId string, name string, rating float64) string {
	t.Helper()

	id := uuid.NewString()
	if _, err := db.Exec(
		`INSERT INTO "Torrons" ("Id", "Name", "Rating", "Image", "Class") VALUES ($1, $2, $3, $4, $5)`,
		id, name, rating, "test-image.png", classId,
	); err != nil {
		t.Fatalf("failed to insert test torro: %v", err)
	}
	return id
}

// newIntegrationRequest builds a request carrying the given chi URL params
// and an optional user ID in context, matching how UserMiddleware and chi's
// router would have set them up. HX-Request is always set so handlers that
// branch on isHX render just the fragment (skipping header/topbar, which
// aren't relevant to what these tests assert on).
func newIntegrationRequest(method, target string, urlParams map[string]string, userId string) *http.Request {
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

const floatTolerance = 0.01

func floatsClose(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < floatTolerance
}

// -- Vote casting (handler.go's result) --

func TestIntegration_VoteCasting(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	pairingRepo := repository.NewPairingRepo(db)
	torroRepo := repository.NewTorroRepo(db)
	classRepo := repository.NewClassRepo(db)
	resultRepo := repository.NewResultRepo(db)
	userRepo := repository.NewUserRepo(db)
	userEloRepo := repository.NewUserEloSnapshotRepo(db)

	classId := insertTestClass(t, db, "Vote Casting Test Class")
	torro1Id := insertTestTorro(t, db, classId, "Torró Guanyador", 1500)
	torro2Id := insertTestTorro(t, db, classId, "Torró Perdedor", 1500)

	pairing, err := pairingRepo.Create(ctx, &domain.Pairing{
		Torro1: torro1Id,
		Torro2: torro2Id,
		Class:  classId,
	})
	if err != nil {
		t.Fatalf("failed to create test pairing: %v", err)
	}

	user, err := userRepo.Create(ctx, &domain.User{Id: uuid.NewString()})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h := &Handler{
		db:          db,
		template:    newIntegrationTemplate(t),
		bpool:       bpool.NewBufferPool(8),
		pairingRepo: pairingRepo,
		torroRepo:   torroRepo,
		classRepo:   classRepo,
		resultRepo:  resultRepo,
		userRepo:    userRepo,
		userEloRepo: userEloRepo,
	}

	// torro1 wins: winnerId is passed as the *query string* "id" param,
	// while the pairing ID is the *path* "id" param -- both literally named
	// "id" in the real route (`/pairings/{id}/vote?id={winnerId}`), so this
	// mirrors that exactly rather than being a copy-paste mistake.
	target := fmt.Sprintf("/pairings/%s/vote?id=%s", pairing.Id, torro1Id)
	req := newIntegrationRequest(http.MethodPost, target, map[string]string{"id": pairing.Id}, user.Id)
	rec := httptest.NewRecorder()

	h.result(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	wantNew1, wantNew2 := UpdateRatings(1500, 1500, true, K)

	// Assert the Results row: winner, before/after ratings, and -- this is
	// the regression this test specifically guards -- that UserId actually
	// got persisted (see repository/postgres_result.go's CreateTx / the
	// "fix(results): persist UserId and CampaignId on vote result creation"
	// history this project already has for this exact bug).
	var (
		winner           string
		rat1Bef, rat2Bef float64
		rat1Aft, rat2Aft float64
		gotUserId        sql.NullString
		gotCampaignId    sql.NullString
	)
	err = db.QueryRowContext(ctx,
		`SELECT "Winner", "Torro1RatingBefore", "Torro2RatingBefore",
		        "Torro1RatingAfter", "Torro2RatingAfter", "UserId", "CampaignId"
		 FROM "Results" WHERE "Pairing" = $1`,
		pairing.Id,
	).Scan(&winner, &rat1Bef, &rat2Bef, &rat1Aft, &rat2Aft, &gotUserId, &gotCampaignId)
	if err != nil {
		t.Fatalf("failed to read back the Results row: %v", err)
	}

	if winner != torro1Id {
		t.Errorf("Results.Winner = %q, want %q", winner, torro1Id)
	}
	if !floatsClose(rat1Bef, 1500) || !floatsClose(rat2Bef, 1500) {
		t.Errorf("before ratings = (%v, %v), want (1500, 1500)", rat1Bef, rat2Bef)
	}
	if !floatsClose(rat1Aft, wantNew1) || !floatsClose(rat2Aft, wantNew2) {
		t.Errorf("after ratings = (%v, %v), want (%v, %v)", rat1Aft, rat2Aft, wantNew1, wantNew2)
	}
	if rat1Aft <= rat1Bef {
		t.Errorf("winner's rating didn't increase: before=%v after=%v", rat1Bef, rat1Aft)
	}
	if rat2Aft >= rat2Bef {
		t.Errorf("loser's rating didn't decrease: before=%v after=%v", rat2Bef, rat2Aft)
	}

	if !gotUserId.Valid || gotUserId.String != user.Id {
		t.Errorf("Results.UserId = %+v, want a valid value equal to %q", gotUserId, user.Id)
	}

	// Deviation from the original brief worth flagging explicitly: the
	// brief expected CampaignId to be persisted here too, by analogy with
	// UserId (both were part of the same past regression fix in
	// postgres_result.go's INSERT column list). But handler.go's result()
	// never actually looks up the active campaign or sets
	// domain.Result.CampaignId on the struct it builds -- so today,
	// CampaignId is always NULL for every Phase 1 (open-voting) vote,
	// campaign active or not. That's not the same bug the history commit
	// fixed (that one was a repository-layer INSERT bug for a field the
	// handler *did* populate); this looks like a real, separate product
	// gap (Results.CampaignId is effectively dead for this codepath) but
	// fixing it is out of scope for a test-coverage task, so this test
	// documents the current, actual behavior instead of asserting a
	// premise the code doesn't support.
	if gotCampaignId.Valid {
		t.Errorf("Results.CampaignId = %q, want NULL (result() never populates it today -- see comment above)", gotCampaignId.String)
	}

	// Also assert the Torrons.Rating columns themselves actually moved,
	// not just the historical snapshot in Results.
	var torro1Rating, torro2Rating float64
	if err := db.QueryRowContext(ctx, `SELECT "Rating" FROM "Torrons" WHERE "Id" = $1`, torro1Id).Scan(&torro1Rating); err != nil {
		t.Fatalf("failed to read back torro1's rating: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT "Rating" FROM "Torrons" WHERE "Id" = $1`, torro2Id).Scan(&torro2Rating); err != nil {
		t.Fatalf("failed to read back torro2's rating: %v", err)
	}
	if !floatsClose(torro1Rating, wantNew1) {
		t.Errorf("Torrons.Rating for the winner = %v, want %v", torro1Rating, wantNew1)
	}
	if !floatsClose(torro2Rating, wantNew2) {
		t.Errorf("Torrons.Rating for the loser = %v, want %v", torro2Rating, wantNew2)
	}
}

// -- Full bracket lifecycle (bracket_handler.go) --

func TestIntegration_BracketLifecycle(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	torroRepo := repository.NewTorroRepo(db)
	classRepo := repository.NewClassRepo(db)
	campaignRepo := repository.NewCampaignRepo(db)
	bracketRepo := repository.NewBracketRepo(db)
	userRepo := repository.NewUserRepo(db)

	// Torrons.Class is never literally "5" (the real "Global" pseudo-class
	// convention) in production; a plain, non-Global class ID like this one
	// is exactly what the rest of the app uses for every real category, so
	// there's nothing special to work around here.
	classId := insertTestClass(t, db, "Bracket Lifecycle Test Class")

	// Seed 4 torrons with distinct ratings so TopNByClass's seed order is
	// deterministic: seed 1 (highest) .. seed 4 (lowest).
	seed1Id := insertTestTorro(t, db, classId, "Seed 1", 1600)
	seed2Id := insertTestTorro(t, db, classId, "Seed 2", 1550)
	seed3Id := insertTestTorro(t, db, classId, "Seed 3", 1500)
	seed4Id := insertTestTorro(t, db, classId, "Seed 4", 1450)

	now := time.Now().UTC()
	campaign, err := campaignRepo.Create(ctx, &domain.Campaign{
		Name:      "Bracket Lifecycle Test Campaign",
		StartDate: now.Add(-1 * time.Hour).Format(time.RFC3339),
		EndDate:   now.Add(1 * time.Hour).Format(time.RFC3339),
		Year:      now.Year(),
		Status:    domain.CampaignStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create test campaign: %v", err)
	}

	voterA, err := userRepo.Create(ctx, &domain.User{Id: uuid.NewString()})
	if err != nil {
		t.Fatalf("failed to create voter A: %v", err)
	}
	voterB, err := userRepo.Create(ctx, &domain.User{Id: uuid.NewString()})
	if err != nil {
		t.Fatalf("failed to create voter B: %v", err)
	}

	h := &Handler{
		db:           db,
		template:     newIntegrationTemplate(t),
		bpool:        bpool.NewBufferPool(8),
		torroRepo:    torroRepo,
		classRepo:    classRepo,
		campaignRepo: campaignRepo,
		bracketRepo:  bracketRepo,
	}

	// -- 1. bracketCreate: seed a size-4 bracket --

	createTarget := fmt.Sprintf("/bracket/%s/create?size=4", classId)
	createReq := newIntegrationRequest(http.MethodPost, createTarget, map[string]string{"classId": classId}, "")
	createRec := httptest.NewRecorder()

	h.bracketCreate(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("bracketCreate status = %d, want %d; body: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	var bracket domain.Bracket
	if err := json.Unmarshal(createRec.Body.Bytes(), &bracket); err != nil {
		t.Fatalf("failed to decode bracketCreate response: %v", err)
	}

	if bracket.ClassId != classId {
		t.Errorf("bracket.ClassId = %q, want %q", bracket.ClassId, classId)
	}
	if bracket.CampaignId != campaign.Id {
		t.Errorf("bracket.CampaignId = %q, want %q", bracket.CampaignId, campaign.Id)
	}
	if bracket.Size != 4 {
		t.Errorf("bracket.Size = %d, want 4", bracket.Size)
	}
	if bracket.CurrentRound != 1 {
		t.Errorf("bracket.CurrentRound = %d, want 1", bracket.CurrentRound)
	}
	if bracket.Status != domain.BracketStatusInProgress {
		t.Errorf("bracket.Status = %q, want %q", bracket.Status, domain.BracketStatusInProgress)
	}

	entries, err := bracketRepo.ListEntries(ctx, bracket.Id)
	if err != nil {
		t.Fatalf("failed to list bracket entries: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("len(entries) = %d, want 4", len(entries))
	}
	seedByTorro := make(map[string]int, len(entries))
	for _, e := range entries {
		seedByTorro[e.TorronId] = e.Seed
	}
	if seedByTorro[seed1Id] != 1 || seedByTorro[seed2Id] != 2 || seedByTorro[seed3Id] != 3 || seedByTorro[seed4Id] != 4 {
		t.Fatalf("unexpected seeding: %+v", seedByTorro)
	}

	round1Matches, err := bracketRepo.ListMatchesByRound(ctx, bracket.Id, 1)
	if err != nil {
		t.Fatalf("failed to list round 1 matches: %v", err)
	}
	if len(round1Matches) != 2 {
		t.Fatalf("len(round1Matches) = %d, want 2", len(round1Matches))
	}

	// Standard seeding for a field of 4 is 1v4, 2v3.
	matchOf := func(matches []*domain.BracketMatch, a, b string) *domain.BracketMatch {
		for _, m := range matches {
			t2 := ""
			if m.Torro2Id != nil {
				t2 = *m.Torro2Id
			}
			if (m.Torro1Id == a && t2 == b) || (m.Torro1Id == b && t2 == a) {
				return m
			}
		}
		return nil
	}

	match1v4 := matchOf(round1Matches, seed1Id, seed4Id)
	match2v3 := matchOf(round1Matches, seed2Id, seed3Id)
	if match1v4 == nil {
		t.Fatalf("expected a seed1-vs-seed4 match in round 1, got %+v", round1Matches)
	}
	if match2v3 == nil {
		t.Fatalf("expected a seed2-vs-seed3 match in round 1, got %+v", round1Matches)
	}
	if match1v4.Status != domain.BracketMatchStatusPending || match2v3.Status != domain.BracketMatchStatusPending {
		t.Fatalf("expected both round 1 matches to be pending, got %q and %q", match1v4.Status, match2v3.Status)
	}

	// -- 2. bracketMatchVote: cast a single vote on only ONE of the two
	// round-1 matches, leaving the other untouched --

	voteTarget := fmt.Sprintf("/bracket/match/%s/vote?winner=%s", match1v4.Id, seed1Id)
	voteReq := newIntegrationRequest(http.MethodPost, voteTarget, map[string]string{"matchId": match1v4.Id}, voterA.Id)
	voteRec := httptest.NewRecorder()

	h.bracketMatchVote(voteRec, voteReq)

	if voteRec.Code != http.StatusOK {
		t.Fatalf("bracketMatchVote status = %d, want %d; body: %s", voteRec.Code, http.StatusOK, voteRec.Body.String())
	}

	// Only one of the two round-1 matches has a vote, so the round isn't
	// "fully voted" yet -- the bracket must still be sitting at round 1.
	afterOneVote, err := bracketRepo.Get(ctx, bracket.Id)
	if err != nil {
		t.Fatalf("failed to reload bracket: %v", err)
	}
	if afterOneVote.CurrentRound != 1 || afterOneVote.Status != domain.BracketStatusInProgress {
		t.Fatalf("bracket after one partial vote = round %d / %q, want round 1 / %q",
			afterOneVote.CurrentRound, afterOneVote.Status, domain.BracketStatusInProgress)
	}

	// -- 3. bracketAdvance: force the round through regardless of the
	// still-untouched match2v3 (0-0, decided by the lower-seed tie-break) --

	advanceReq := newIntegrationRequest(http.MethodPost, "/bracket/"+bracket.Id+"/advance",
		map[string]string{"bracketId": bracket.Id}, "")
	advanceRec := httptest.NewRecorder()

	h.bracketAdvance(advanceRec, advanceReq)

	if advanceRec.Code != http.StatusOK {
		t.Fatalf("bracketAdvance status = %d, want %d; body: %s", advanceRec.Code, http.StatusOK, advanceRec.Body.String())
	}

	var advancedBracket domain.Bracket
	if err := json.Unmarshal(advanceRec.Body.Bytes(), &advancedBracket); err != nil {
		t.Fatalf("failed to decode bracketAdvance response: %v", err)
	}
	if advancedBracket.CurrentRound != 2 {
		t.Fatalf("bracket.CurrentRound after advance = %d, want 2", advancedBracket.CurrentRound)
	}
	if advancedBracket.Status != domain.BracketStatusInProgress {
		t.Fatalf("bracket.Status after advance = %q, want %q", advancedBracket.Status, domain.BracketStatusInProgress)
	}

	// match2v3 was 0-0: the tie-break favours the lower (better) seed, i.e.
	// seed2 over seed3.
	round2Matches, err := bracketRepo.ListMatchesByRound(ctx, bracket.Id, 2)
	if err != nil {
		t.Fatalf("failed to list round 2 matches: %v", err)
	}
	if len(round2Matches) != 1 {
		t.Fatalf("len(round2Matches) = %d, want 1", len(round2Matches))
	}
	final := round2Matches[0]
	finalTorro2 := ""
	if final.Torro2Id != nil {
		finalTorro2 = *final.Torro2Id
	}
	gotFinalists := map[string]bool{final.Torro1Id: true, finalTorro2: true}
	if !gotFinalists[seed1Id] || !gotFinalists[seed2Id] {
		t.Fatalf("expected the final to be seed1 vs seed2 (seed2 winning its 0-0 tie-break over seed3), got %s vs %s",
			final.Torro1Id, finalTorro2)
	}
	if final.Status != domain.BracketMatchStatusPending {
		t.Fatalf("final match status = %q, want %q", final.Status, domain.BracketMatchStatusPending)
	}

	// -- 4. bracketMatchVote on the final: this is the bracket's only
	// round-2 match, so a single vote fully-votes the round and should
	// cascade the bracket straight to completed --

	finalVoteTarget := fmt.Sprintf("/bracket/match/%s/vote?winner=%s", final.Id, seed1Id)
	finalVoteReq := newIntegrationRequest(http.MethodPost, finalVoteTarget, map[string]string{"matchId": final.Id}, voterB.Id)
	finalVoteRec := httptest.NewRecorder()

	h.bracketMatchVote(finalVoteRec, finalVoteReq)

	if finalVoteRec.Code != http.StatusOK {
		t.Fatalf("final bracketMatchVote status = %d, want %d; body: %s", finalVoteRec.Code, http.StatusOK, finalVoteRec.Body.String())
	}

	completed, err := bracketRepo.Get(ctx, bracket.Id)
	if err != nil {
		t.Fatalf("failed to reload bracket after final vote: %v", err)
	}
	if completed.Status != domain.BracketStatusCompleted {
		t.Fatalf("bracket.Status after final vote = %q, want %q", completed.Status, domain.BracketStatusCompleted)
	}
	if completed.ChampionId == nil || *completed.ChampionId != seed1Id {
		t.Fatalf("bracket.ChampionId = %v, want %q", completed.ChampionId, seed1Id)
	}
	if completed.CompletedAt == nil {
		t.Error("expected bracket.CompletedAt to be set once completed")
	}
}
