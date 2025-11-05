package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

func New(c *config.Config) *Torrons {
	db, err := NewDatabaseConnection(c.Database)
	if err != nil {
		logger.Fatal("[API - New] - "+
			"Failed to connect to database. %v", err)
	}

	// A buffer pool is created to safely check template
	// execution and properly handle the errors
	bpool := bpool.NewBufferPool(64)

	paringRepo := repository.NewPairingRepo(db)
	torroRepo := repository.NewTorroRepo(db)
	classRepo := repository.NewClassRepo(db)
	resultRepo := repository.NewResultRepo(db)

	if err := CheckPairingsCreated(paringRepo, torroRepo, classRepo); err != nil {
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
	)
	srv := http.New(c.Port, handler)

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
	t.srv.Shutdown()
}

func CheckPairingsCreated(
	pairingRep domain.PairingRepo,
	torroRep domain.TorroRepo,
	classRep domain.ClassRepo,
) error {
	classes, err := classRep.List()
	if err != nil {
		return err
	}

	for _, c := range classes {
		count, err := pairingRep.CountClass(c.Id)
		if err != nil {
			return err
		}

		if count < 1 {
			logger.Info("[API - New] - "+
				"%s - Creating pairings", c.Name)

			t, err := torroRep.ListByClass(c.Id)
			if err != nil {
				return err
			}

			count := 0
			for i := 0; i < len(t); i++ {
				for j := i + 1; j < len(t); j++ {
					pairing := domain.Pairing{
						Torro1: t[i].Id,
						Torro2: t[j].Id,
						Class:  c.Id,
					}

					_, err := pairingRep.Create(&pairing)
					if err != nil {
						logger.Error("[API - New] - "+
							"%s - Failed to create pairing (%s vs %s). %v",
							c.Name,
							t[i].Id,
							t[j].Id,
							err)
						continue
					}
					count++
				}
			}

			logger.Info("[API - New] - "+
				"%s - Created %d pairings", c.Name, count)
			continue
		}

		logger.Info("[API - New] - "+
			"%s - %d pairings already created", c.Name, count)
	}

	return nil
}

func NewDatabaseConnection(c config.Database) (*sql.DB, error) {
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
	dbConnection.SetMaxOpenConns(25)                        // Maximum total connections (in-use + idle)
	dbConnection.SetMaxIdleConns(5)                         // Maximum idle connections in pool
	dbConnection.SetConnMaxLifetime(5 * time.Minute)        // Maximum connection age
	dbConnection.SetConnMaxIdleTime(1 * time.Minute)        // Maximum idle time before close

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
