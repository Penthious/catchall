package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	MaxIdleTime  time.Duration
	DisableTLS   bool
}

// Open opens a connection to the database.
func Open(cfg Config) (*bun.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	pgDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(u.String())))
	pgDB.SetMaxOpenConns(cfg.MaxOpenConns)
	pgDB.SetMaxIdleConns(cfg.MaxIdleConns)
	pgDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	db := bun.NewDB(pgDB, pgdialect.New())

	return db, nil
}

// StatusCheck tries to ping the database to make sure it is up. It will then make a full round trip connection
// via a `select true;` query to make sure the database is ready to accept connections.
func StatusCheck(ctx context.Context, pgdb *bun.DB) error {

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = pgdb.DB.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return fmt.Errorf("failed to ping postgres db: %w", ctx.Err())
		}
	}

	// Make sure we didn't time out or be cancelled.
	if ctx.Err() != nil {
		return fmt.Errorf("failed on timeout db: %w", ctx.Err())
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return pgdb.QueryRowContext(ctx, q).Scan(&tmp)
}
