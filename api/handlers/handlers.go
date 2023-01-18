package handlers

import (
	v1 "github.com/penthious/catchall/api/handlers/v1"
	"github.com/penthious/catchall/business/ports"
	"github.com/penthious/catchall/foundation/web"
	"github.com/penthious/catchall/foundation/web/middleware"
	"github.com/rs/zerolog"
	"net/http"
	"os"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Log         *zerolog.Logger
	DB          ports.DB
	ServiceName string
	Shutdown    chan os.Signal
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler {

	// the middleware forms a stack (FIFO)
	app := web.NewApp(
		cfg.ServiceName,
		cfg.Shutdown,
		cfg.Log,
		middleware.LogRequest(),
		middleware.Errors(), // after this point any errors will be lost
		// Any extra middleware can be added here:
		// cors
		// ratelimiter
		// prom metrics
		// panic recovery
	)

	v1.Routes(
		app,
		v1.Options{
			DB: cfg.DB,
		},
	)

	return app
}
