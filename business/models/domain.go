package models

import "github.com/uptrace/bun"

// Domain is the domain model.
// TODO: remove bun dependency, only here to make it easier to create the table (see adapters/postgres_repo.go)
type Domain struct {
	bun.BaseModel `bun:"table:domains,alias:d"`
	ID            int64 `bun:",pk,autoincrement"`

	Domain    string `bun:",unique"`
	Bounced   int
	Delivered int
}
