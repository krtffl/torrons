package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/config"
	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/http"
	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/internal/repository"
)

type Torrons struct {
	cfg *config.Config
	srv *http.Server
	db  *sql.DB

	eCh chan error
}

func New(c *config.Config, runMigrations bool) *Torrons {
	db, err := NewDatabaseConnection(c.Database, runMigrations)
	if err != nil {
		logger.Fatal("[API - New] - "+
			"Failed to connect to database. %v", err)
	}

	// A buffer pool is created to safely check template
	// execution and properly handle the errors
	// Pool size of 64 balances memory usage with concurrency:
	// - Allows up to 64 concurrent template renderings
	// - Reduces GC pressure by reusing buffers
	// - Typical size for applications with moderate concurrent requests
	bpool := bpool.NewBufferPool(64)

	paringRepo := repository.NewPairingRepo(db)
	torroRepo := repository.NewTorroRepo(db)
	classRepo := repository.NewClassRepo(db)
	resultRepo := repository.NewResultRepo(db)
	userRepo := repository.NewUserRepo(db)
	userEloRepo := repository.NewUserEloSnapshotRepo(db)
	campaignRepo := repository.NewCampaignRepo(db)
	bracketRepo := repository.NewBracketRepo(db)
	adventVoteRepo := repository.NewAdventVoteRepo(db)
	friendCircleRepo := repository.NewFriendCircleRepo(db)
	pressStatsRepo := repository.NewPressStatsRepo(db)
	wrappedStatsRepo := repository.NewWrappedStatsRepo(db)
	personaRepo := repository.NewPersonaRepo(db)

	if err := CheckPairingsCreated(db, paringRepo, torroRepo, classRepo); err != nil {
		logger.Fatal("[API - New] - "+
			"Failed to check pairings. %v", err)
	}

	handler := http.NewHandler(
		db,
		bpool,
		paringRepo,
		torroRepo,
		classRepo,
		resultRepo,
		userRepo,
		userEloRepo,
		campaignRepo,
		bracketRepo,
		adventVoteRepo,
		friendCircleRepo,
		pressStatsRepo,
		wrappedStatsRepo,
		personaRepo,
		c.AdminToken,
	)

	if c.AdminToken == "" {
		logger.Warn("[API - New] ADMIN_TOKEN is not set - bracket admin endpoints (create/advance) will reject all requests")
	}

	srv := http.New(c.Port, handler, c.TrustedProxies)

	return &Torrons{
		cfg: c,
		srv: srv,
		db:  db,
		eCh: make(chan error, 1),
	}
}

func (t *Torrons) Run() {
	go func() { t.eCh <- t.srv.Run() }()
	func() {
		for err := range t.eCh {
			if err != nil {
				logger.Fatal("[API - Run] - Couldn't run API. %v", err)
			}
		}
	}()
}

func (t *Torrons) Shutdown() {
	logger.Info("[API - Shutdown] Shutting down gracefully")

	// Shutdown HTTP server first (stop accepting new requests)
	t.srv.Shutdown()

	// Close database connection
	if err := t.db.Close(); err != nil {
		logger.Error("[API - Shutdown] Failed to close database connection. %v", err)
	} else {
		logger.Info("[API - Shutdown] Database connection closed")
	}

	logger.Info("[API - Shutdown] Shutdown complete")
}

// createGlobalPairings creates strategic pairings for the global category
// Instead of all O(n²) combinations, this creates O(n*k) pairings where:
// - Each torron is compared with top torrons from other categories
// - This allows ELO to converge to a global ranking efficiently
func createGlobalPairings(
	ctx context.Context,
	tx *sql.Tx,
	torroRep domain.TorroRepo,
	globalClassId string,
) (int, error) {
	// Get torrons from each category (1-4), sorted by rating (best first)
	categories := []string{"1", "2", "3", "4"}
	torronsByCategory := make(map[string][]*domain.Torro)

	for _, categoryId := range categories {
		torrons, err := torroRep.ListByClass(ctx, categoryId)
		if err != nil {
			return 0, err
		}
		// Filter out discontinued torrons
		var activeTorrons []*domain.Torro
		for _, t := range torrons {
			if !t.Discontinued {
				activeTorrons = append(activeTorrons, t)
			}
		}
		torronsByCategory[categoryId] = activeTorrons
	}

	pairingsCreated := 0
	pairingsMap := make(map[string]bool) // Prevent duplicates

	// Strategy: Each torron competes with representatives from other categories
	// For thorough comparison, pair each torron with:
	// - Top 5 from each other category (if available)
	// - This ensures cross-category comparisons while keeping count manageable
	const pairingsPerCategory = 5

	for categoryId, torrons := range torronsByCategory {
		for _, torron := range torrons {
			// Pair this torron with top torrons from OTHER categories
			for otherCategoryId, otherTorrons := range torronsByCategory {
				if otherCategoryId == categoryId {
					continue // Skip same category
				}

				// Pair with top N torrons from this other category
				limit := pairingsPerCategory
				if len(otherTorrons) < limit {
					limit = len(otherTorrons)
				}

				for i := 0; i < limit; i++ {
					otherTorron := otherTorrons[i]

					// Create pairing key to prevent duplicates (order-independent)
					var pairingKey string
					if torron.Id < otherTorron.Id {
						pairingKey = torron.Id + ":" + otherTorron.Id
					} else {
						pairingKey = otherTorron.Id + ":" + torron.Id
					}

					// Skip if already created
					if pairingsMap[pairingKey] {
						continue
					}

					pairing := domain.Pairing{
						Torro1: torron.Id,
						Torro2: otherTorron.Id,
						Class:  globalClassId,
					}

					inserted, err := insertPairingTx(ctx, tx, &pairing)
					if err != nil {
						return 0, fmt.Errorf("create global pairing (%s vs %s): %w",
							torron.Id, otherTorron.Id, err)
					}

					pairingsMap[pairingKey] = true
					if inserted {
						pairingsCreated++
					}
				}
			}
		}
	}

	return pairingsCreated, nil
}

// insertPairingTx inserts a single pairing within a transaction, silently
// skipping it if the same matchup (order-independent) already exists for this
// class - see migration idx_pairings_unique_matchup. This makes concurrent
// seeding across two instances that both raced to seed the same class safe:
// whichever loses the row-level conflict just no-ops instead of erroring or
// double-inserting. Returns whether a row was actually inserted, so callers
// can report an accurate created-count even when some inserts are skipped.
func insertPairingTx(ctx context.Context, tx *sql.Tx, p *domain.Pairing) (bool, error) {
	var id string
	err := tx.QueryRowContext(ctx,
		`INSERT INTO "Pairings" ("Id", "Torro1", "Torro2", "Class")
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (LEAST("Torro1", "Torro2"), GREATEST("Torro1", "Torro2"), "Class") DO NOTHING
		 RETURNING "Id"`,
		uuid.NewString(), p.Torro1, p.Torro2, p.Class,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // matchup already exists - not an error
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// createAllPairs seeds every unordered pair of a regular class's torrons within
// the given transaction.
func createAllPairs(ctx context.Context, tx *sql.Tx, torroRep domain.TorroRepo, classId string) (int, error) {
	t, err := torroRep.ListByClass(ctx, classId)
	if err != nil {
		return 0, err
	}

	created := 0
	for i := 0; i < len(t); i++ {
		for j := i + 1; j < len(t); j++ {
			pairing := domain.Pairing{Torro1: t[i].Id, Torro2: t[j].Id, Class: classId}
			inserted, err := insertPairingTx(ctx, tx, &pairing)
			if err != nil {
				return 0, fmt.Errorf("create pairing (%s vs %s): %w", t[i].Id, t[j].Id, err)
			}
			if inserted {
				created++
			}
		}
	}
	return created, nil
}

func CheckPairingsCreated(
	db *sql.DB,
	pairingRep domain.PairingRepo,
	torroRep domain.TorroRepo,
	classRep domain.ClassRepo,
) error {
	ctx := context.Background()
	classes, err := classRep.List(ctx)
	if err != nil {
		return err
	}

	for _, c := range classes {
		count, err := pairingRep.CountClass(ctx, c.Id)
		if err != nil {
			return err
		}

		if count >= 1 {
			logger.Info("[API - New] - %s - %d pairings already created", c.Name, count)
			continue
		}

		logger.Info("[API - New] - %s - Creating pairings", c.Name)

		// Seed a class's pairings inside a single transaction: if any insert
		// fails the whole class rolls back, instead of leaving a permanently
		// partial set that the count<1 guard above would then treat as "done".
		// One commit per class is also far faster than the previous per-row
		// autocommit (thousands of round-trips).
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		var created int
		if c.Id == "5" {
			// Global (cross-category) uses the smart pairing strategy.
			created, err = createGlobalPairings(ctx, tx, torroRep, c.Id)
		} else {
			created, err = createAllPairs(ctx, tx, torroRep, c.Id)
		}
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		logger.Info("[API - New] - %s - Created %d pairings", c.Name, created)
	}

	return nil
}

func NewDatabaseConnection(c config.Database, runMigrations bool) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode,
	)

	// creates the connection but does not validate it
	dbConnection, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// validates the connection is ok
	if err := dbConnection.Ping(); err != nil {
		return nil, err
	}

	// Configure connection pool to prevent resource exhaustion
	dbConnection.SetMaxOpenConns(25)                 // Maximum total connections (in-use + idle)
	dbConnection.SetMaxIdleConns(5)                  // Maximum idle connections in pool
	dbConnection.SetConnMaxLifetime(5 * time.Minute) // Maximum connection age
	dbConnection.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time before close

	// This embedded run is the single migration mechanism. Honor the
	// skip-migrations flag here (previously it ran unconditionally, so the flag
	// was a no-op).
	if !runMigrations {
		logger.Info("[API - NewDatabaseConnection] - Skipping migrations (--skip-migrations)")
		return dbConnection, nil
	}

	driver, err := postgres.WithInstance(dbConnection, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	// migration files are embedded
	d, err := iofs.New(torrons.Migrations, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		"postgres",
		driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && err == migrate.ErrNoChange {
		v, _, err := m.Version()
		if err != nil {
			return nil, err
		}

		logger.Info(
			"[API - NewDatabaseConnection] "+
				" - Database already at latest version: %d",
			v,
		)

		return dbConnection, nil
	}
	if err != nil {
		return nil, err
	}

	newV, _, err := m.Version()
	if err != nil {
		return nil, err
	}

	logger.Info("[API - NewDatabaseConnection] "+
		"- Migrated database to version %d", newV)

	return dbConnection, nil
}
