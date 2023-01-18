// Package domain_grp maintains the group of handlers.
package domain_grp

import (
	"fmt"
	"github.com/penthious/catchall/business/ports"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mailgun/catchall"
	"github.com/penthious/catchall/foundation/web"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	DB ports.DB
}

// Get queries the database for a domain and returns the different types of catchall events.
func (h Handlers) Get(ctx echo.Context) error {
	domainName := ctx.Param("domain_name")

	domain, err := h.DB.Query(domainName)
	if err != nil {
		return fmt.Errorf("error getting domain: %w", err)
	}

	if domain.Bounced > 0 {
		return web.Respond(ctx, http.StatusOK, "not catch-all")
	}

	if domain.Delivered >= 1_000 {
		return web.Respond(ctx, http.StatusOK, "catch-all")
	}

	return web.Respond(ctx, http.StatusOK, "unknown")
}

// PutDelivered updates the delivered count for a domain.
func (h Handlers) PutDelivered(ctx echo.Context) error {
	event := catchall.Event{
		Type:   catchall.TypeDelivered,
		Domain: ctx.Param("domain_name"),
	}

	if err := h.DB.Insert(event); err != nil {
		return fmt.Errorf("error saving delivered: %w", err)
	}

	return web.Respond(ctx, http.StatusNoContent, nil)
}

// PutBounced updates the bounced count for a domain.
func (h Handlers) PutBounced(ctx echo.Context) error {
	event := catchall.Event{
		Type:   catchall.TypeBounced,
		Domain: ctx.Param("domain_name"),
	}

	if err := h.DB.Insert(event); err != nil {
		return fmt.Errorf("error saving bounced: %w", err)
	}

	return web.Respond(ctx, http.StatusNoContent, nil)
}
