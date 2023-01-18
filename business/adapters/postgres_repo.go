package adapters

import (
	"context"
	"fmt"
	"github.com/mailgun/catchall"
	"github.com/penthious/catchall/business/models"
	"github.com/penthious/catchall/business/ports"
	webErr "github.com/penthious/catchall/foundation/web/errors"
	"github.com/uptrace/bun"
	"strings"
)

var _ ports.DB = PostgresRepo{}

// NewPostgresRepo returns a new PostgresRepo.
func NewPostgresRepo(db *bun.DB) PostgresRepo {

	// Create domain table if it doesn't exist
	// TODO: Move this to a migration via goose or something
	_, err := db.NewCreateTable().Model((*models.Domain)(nil)).Exec(context.Background())
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		// panicing here because we can't continue without the table, and we cant handle an error here unless we
		// change the function signature which is not a wanted change (this panic will be removed when
		// migrations are setup)
		panic(err)
	}

	return PostgresRepo{db: db}
}

// PostgresRepo manages the set of API's for domain data.
type PostgresRepo struct{ db *bun.DB }

// Query returns the domain for the given domain name.
func (p PostgresRepo) Query(domain string) (models.Domain, error) {
	var d models.Domain
	if err := p.db.NewSelect().Model(&models.Domain{}).Where("domain = ?", domain).Scan(context.Background(), &d); err != nil {
		return models.Domain{}, fmt.Errorf("error querying domain: %w", err)
	}
	return d, nil
}

// Insert inserts the given event into the database.
func (p PostgresRepo) Insert(event catchall.Event) error {

	d, err := p.Query(event.Domain)
	if webErr.IsNoRowsError(err) {
		d.Domain = event.Domain
		switch event.Type {
		case catchall.TypeBounced:
			d.Bounced = 1
		case catchall.TypeDelivered:
			d.Delivered = 1
		}

		if _, err := p.db.NewInsert().Model(&d).Exec(context.Background()); err != nil {
			return fmt.Errorf("error inserting domain: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("error querying domain: %w", err)
	}

	switch event.Type {
	case catchall.TypeBounced:
		d.Bounced++
	case catchall.TypeDelivered:
		d.Delivered++
	}

	if _, err := p.db.NewUpdate().Model(&d).Where("domain = ?", event.Domain).Exec(context.Background()); err != nil {
		return fmt.Errorf("error updating domain: %w", err)
	}

	return nil
}
