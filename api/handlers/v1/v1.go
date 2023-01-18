package v1

import (
	"github.com/penthious/catchall/api/handlers/v1/domain_grp"
	"github.com/penthious/catchall/business/ports"
	"github.com/penthious/catchall/foundation/web"
	"net/http"
)

const v1 = "v1"

type Options struct {
	DB ports.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Options) {
	dgrp := domain_grp.Handlers{
		DB: cfg.DB,
	}
	app.Handle(http.MethodGet, v1, "/domain/:domain_name", dgrp.Get)
	app.Handle(http.MethodPut, v1, "/events/:domain_name/bounced", dgrp.PutBounced)
	app.Handle(http.MethodPut, v1, "/events/:domain_name/delivered", dgrp.PutDelivered)
}
