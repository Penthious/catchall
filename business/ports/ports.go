package ports

import (
	"github.com/mailgun/catchall"
	"github.com/penthious/catchall/business/models"
)

// DB defines the interface for the database.
type DB interface {
	Query(domain string) (models.Domain, error)
	Insert(event catchall.Event) error
}
