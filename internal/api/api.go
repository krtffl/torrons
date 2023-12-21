package api

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/config"
	"github.com/krtffl/torro/internal/http"
	"github.com/krtffl/torro/internal/logger"
)

type Torrons struct {
	cfg *config.Config
	srv *http.Server

	eCh chan error
}

func New(c *config.Config) *Torrons {
	_, err := NewDatabaseConnection(c.Database)
	if err != nil {
		logger.Fatal("[API - New] - "+
			"Failed to connect to database. %v", err)
	}

	// A buffer pool is created to safely check template
	// execution and properly handle the errors
	bpool := bpool.NewBufferPool(64)
	handler := http.NewHandler(bpool)

	srv := http.New(c.Port, handler)

	return &Torrons{
		cfg: c,
		srv: srv,
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

func NewDatabaseConnection(c config.Database) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
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
